# Air780 Lua 脚本优化建议

## 1. 内存管理优化：从 String 拼接改为 Table 缓冲

### 为什么 `..` 拼接会有问题？

在 Lua 中，**字符串是不可变的（Immutable）**。
这意味着当你执行 `buff = buff .. chunk` 时，Lua 并不是在原有的 `buff` 后面追加数据，而是：
1.  分配一块新的内存，大小 = `len(buff) + len(chunk)`。
2.  把旧的 `buff` 内容复制过去。
3.  把 `chunk` 内容复制过去。
4.  将旧的 `buff` 标记为垃圾等待回收。

**后果：**
如果是高频接收串口数据（例如大数据包被切分成很多小片段到达），会触发**O(N²)**级别的内存复制和大量的内存碎片。这会导致：
*   **GC 频繁**：系统被迫频繁停止工作来清理垃圾，导致 CPU 占用飙升。
*   **内存震荡**：在内存较小的模块（Air780 可用内存通常只有几百KB）上，很容易申请不到大块连续内存而报错重启。

### 为什么 `table.insert` 更好？

Lua 的 `table`（表）在底层通常是预分配内存的数组。
*   `table.insert(t, chunk)` 只是将 `chunk` 的引用放入数组，**不发生字符串复制**。
*   只有当你最终调用 `table.concat(t)` 时，Lua 才会一次性计算总长度并分配内存，将所有片段拼起来。

### 代码改造实现

**原代码 (`main.lua`)：**
```lua
uart_recv_buffer = uart_recv_buffer .. chunk
-- 后面直接用 uart_recv_buffer:find()
```

**优化后代码：**

```lua
-- 定义一个 table 作为缓冲区
local uart_buffer_table = {}

uart.on(uartid, "receive", function(id, len)
    -- 1. 读取新数据
    local chunk = uart.read(id, len)
    if not chunk or #chunk == 0 then return end

    -- 2. 存入 table (零拷贝)
    table.insert(uart_buffer_table, chunk)

    -- 3. 策略：为了兼顾性能，我们不每次都 concat，
    --    而是判断缓冲区累计长度，或者积累一定次数后再处理
    --    这里简单起见，演示如何合并检查

    -- 将当前所有片段合并为字符串进行处理
    local current_str = table.concat(uart_buffer_table)

    -- 查找包头 CMD_START
    local start_pos = current_str:find("CMD_START:", 1, true)
    if not start_pos then
        -- 如果缓冲区太大（例如超过4K）且没找到头，说明全是垃圾数据，清理掉
        if #current_str > CONFIG.UART_BUFFER_MAX then
            uart_buffer_table = {} -- 清空
        end
        return
    end

    -- 查找包尾 :CMD_END
    local end_pos = current_str:find(":CMD_END", start_pos + 10, true)

    if end_pos then
        -- 找到完整包！
        local json_str = current_str:sub(start_pos + 10, end_pos - 1)

        -- 处理业务逻辑...
        local success, cmd = pcall(json.decode, json_str)
        if success then process_uart_command(cmd) end

        -- 【关键步骤】清理缓冲区
        -- 保留 end_pos 之后的数据（粘包处理），放入新的 table
        local left_over = current_str:sub(end_pos + 8)
        uart_buffer_table = {}
        if #left_over > 0 then
            table.insert(uart_buffer_table, left_over)
        end
    end
end)
```

## 2. 消息队列机制：从并行争抢改为串行排队

### 为什么目前的 `sys.taskInit` 有风险？

目前的实现是：
```lua
if cmd_data.action == "send_sms" then
    sys.taskInit(function()
        sms.sendLong(...).wait(...)
    end)
end
```

**风险点：**
`sys.taskInit` 会启动一个新的协程（Coroutine）。如果你在一秒内连续发了 5 条短信指令：
1.  Lua 会瞬间启动 5 个协程。
2.  这 5 个协程几乎同时向底层的 4G 模组发送 AT 指令。
3.  **但是，AT 指令通道通常是单工的！** 模组正在处理第1条短信发送（需要几秒），此时第2条指令插进来，底层只能返回 `BUSY` 错误，或者直接丢弃，甚至导致模组死锁。

这就是所谓的“并发争抢硬件资源”。

### 消息队列模式（Producer-Consumer）

我们需要一个**队列（Queue）**来缓冲请求，和一个**单线程消费者（Worker）**来逐个执行。

*   **生产者**：UART 接收到指令，只负责把任务扔进队列，立即返回。
*   **消费者**：一个永远在后台运行的循环，发现队列有任务就拿出来发，发完一个再发下一个。

### 代码改造实现

**第一步：定义队列和消费者**

在 `main.lua` 的全局变量区域添加：

```lua
-- 发送任务队列
local sms_send_queue = {}

-- 启动一个专门的后台协程（消费者）
sys.taskInit(function()
    while true do
        -- 1. 判断队列是否有任务
        if #sms_send_queue > 0 then
            -- 2. 取出第一个任务 (FIFO)
            local task = table.remove(sms_send_queue, 1)

            log.info("Queue", "开始处理短信任务", task.to, "剩余任务:", #sms_send_queue)

            -- 3. 执行发送（阻塞操作，独占硬件）
            local result = sms.sendLong(task.to, task.content).wait(CONFIG.SMS_SEND_TIMEOUT)

            -- 4. 发送结果回传给上位机
            send_to_uart({
                type = "sms_send_result",
                success = (result == true),
                request_id = task.request_id,
                to = task.to
            })

            -- 5. 适当延时，给模组喘息时间（可选，但建议加）
            sys.wait(500)
        else
            -- 队列为空，挂起等待新消息通知，或者简单轮询
            sys.waitUntil("NEW_SMS_TASK", 1000)
        end
    end
end)
```

**第二步：改造 UART 命令处理（生产者）**

修改 `process_uart_command` 函数：

```lua
function process_uart_command(cmd_data)
    -- ... 前面代码不变 ...

    if cmd_data.action == "send_sms" then
        -- 仅仅是将任务入队，不执行耗时操作
        local task = {
            to = cmd_data.to,
            content = cmd_data.content,
            request_id = cmd_data.request_id or os.time()
        }

        table.insert(sms_send_queue, task)

        -- 唤醒消费者协程
        sys.publish("NEW_SMS_TASK")

        log.info("CMD", "短信任务已入队", task.to)

    -- ... 其他代码不变 ...
    end
end
```

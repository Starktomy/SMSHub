-- =================================================================================
-- PROJECT: UART SMS Forwarder
-- DEVICE:  Air780EHV
-- VERSION: 1.2.0 (Dual UART + Optimization)
-- 协议说明：
--   上行（MCU -> 模块）：CMD_START:{json}:CMD_END
--   下行（模块 -> MCU）：SMS_START:{json}:SMS_END
-- =================================================================================

PROJECT = "smshub"
VERSION = "1.2.0"

-- 配置参数
local CONFIG = {
    HEARTBEAT_INTERVAL = 60000,  -- 心跳间隔（毫秒）
    SMS_SEND_TIMEOUT = 30000,    -- 短信发送超时（毫秒）
    MAX_BUFFER_SIZE = 50,        -- 消息缓冲区大小
    UART_BUFFER_MAX = 4096,      -- UART接收缓冲区最大值
}

log.info("main", PROJECT, VERSION)

-- 1. 引入必要库
sys = require("sys")

-- 2. 全局配置与变量
-- 定义两个串口：USB虚拟串口 和 物理串口(Main UART)
local uart_usb = uart.VUART_0 or 0  -- 默认USB是0，防止为nil
local uart_phy = uart.UART_1 or 1   -- 默认物理口是1

-- 为每个串口维护独立的接收缓冲区 (使用 table 优化内存)
local uart_buffers = {}
-- 初始化缓冲区，防止 key 为 nil 导致崩溃
if uart_usb then uart_buffers[uart_usb] = {} end
if uart_phy then uart_buffers[uart_phy] = {} end

-- LED 控制
-- 已移除 LED 控制逻辑


-- 短信发送任务队列
local sms_send_queue = {}

local msg_buffer = {}
local call_ring_count = 0  -- 来电响铃计数

-- ========== 关键：禁用自动数据连接 ==========
mobile.setAuto(0)

-- 3. 看门狗
if wdt then
    wdt.init(9000)
    sys.timerLoopStart(wdt.feed, 3000)
end

-- 初始化两个串口
uart.setup(uart_usb, 115200, 8, 1)
uart.setup(uart_phy, 115200, 8, 1)
log.info("System", "双串口初始化成功 (USB & PHY)")

-- =================================================================================
-- 工具函数区
-- =================================================================================

function get_mobile_info()
    local info = {}
    -- 使用 status 判断：0=未注册 1=已注册 2=搜索中 3=拒绝 5=漫游注册
    local net_stat = mobile.status()
    local iccid = mobile.iccid()
    info.sim_ready = (iccid ~= nil and iccid ~= "" and iccid ~= "unknown")
    info.iccid = iccid or "unknown"
    info.imsi = mobile.imsi() or "unknown"
    info.imei = mobile.imei() or "unknown"  -- 设备唯一标识
    info.number = mobile.number(0) or ""  -- 获取手机号，可能为空

    -- 获取信号强度指标
    local csq = mobile.csq() or 0 -- 范围 0-31，越大越好
    info.csq = csq
    info.rssi = mobile.rssi() or -113  -- 范围 0到-114，值越大越好
    info.rsrp = mobile.rsrp() or -140  -- 范围 -44到-140，值越大越好 (4G模块)
    info.rsrq = mobile.rsrq() or -20   -- 范围 -3到-19.5，值越大越好 (4G模块)

    -- 根据 CSQ 判断信号等级（仅供参考，4G模块应参考rsrp/rsrq）
    if csq == 0 or csq == 99 then
        info.signal_level = 0
        info.signal_desc = "无信号"
    else
        info.signal_level = csq
        info.signal_desc = csq >= 20 and "强" or (csq >= 10 and "中" or "弱")
    end

    info.is_registered = (net_stat == 1 or net_stat == 5)
    info.is_roaming = net_stat == 5
    info.uptime = mcu.ticks2() -- 单位为秒

    -- 获取实时网络信息 (MCC/MNC)
    local net_info = mobile.getNetInfo()
    if net_info and net_info.mcc and net_info.mnc then
        -- 将 mcc/mnc 统一转换为 460/01 这种十进制字符串并拼接
        -- 这里的 mcc 和 mnc 在不同固件下可能是 number (hex) 或 number (dec)
        -- 我们直接将其转为 16 进制字符串再转回 10 进制是 LuatOS 处理 MCC/MNC 的常见技巧
        local mcc_str = string.format("%03X", net_info.mcc)
        local mnc_str = string.format("%02X", net_info.mnc)
        -- 港澳台及国外 MNC 可能是 3 位，这里暂按 2 位处理，后端做模糊匹配
        info.mnc = mcc_str .. mnc_str
    end

    -- https://docs.openluat.com/osapi/core/mobile/#mobileflymodeindex-enable
    info.flymode = mobile.flymode()

    return info
end

-- 广播发送到所有串口
function send_to_uart(data)
    local ok, json_str = pcall(json.encode, data)
    if ok and json_str then
        local packet = "SMS_START:" .. json_str .. ":SMS_END\r\n"
        -- 同时写给两个端口
        uart.write(uart_usb, packet)
        uart.write(uart_phy, packet)
        return true
    else
        log.error("UART", "JSON Encode Failed", json_str)
        return false
    end
end

function process_uart_command(cmd_data)
    if not cmd_data.action then
        send_to_uart({type = "error", msg = "missing action"})
        return
    end

    if cmd_data.action == "send_sms" and cmd_data.to and cmd_data.content then
        -- 优化：将发送任务入队，而不是直接发送
        local task = {
            to = cmd_data.to,
            content = cmd_data.content,
            request_id = cmd_data.request_id or os.time()
        }
        table.insert(sms_send_queue, task)
        sys.publish("NEW_SMS_TASK") -- 唤醒消费者
        log.info("CMD", "短信任务已入队", task.to)

    elseif cmd_data.action == "get_status" then
        send_to_uart({
            type = "status_response",
            timestamp = os.time(),
            mem_kb = math.floor(collectgarbage("count")),
            version = VERSION,
            mobile = get_mobile_info()
        })

    elseif cmd_data.action == "set_flymode" and cmd_data.enabled ~= nil then
        -- 规范化为布尔值
        local flymode_enabled = (cmd_data.enabled == true or cmd_data.enabled == 1 or
                                 cmd_data.enabled == "true" or cmd_data.enabled == "1")

        mobile.flymode(0, flymode_enabled)

        if not flymode_enabled then
            mobile.setAuto(0) -- 退出飞行模式后，确保不自动拨号
        end

        send_to_uart({
            type = "cmd_response",
            action = "set_flymode",
            result = "ok"
        })

    elseif cmd_data.action == "reset_stack" then
        log.info("CMD", "重启协议栈")
        mobile.reset()
        mobile.setAuto(0)
        send_to_uart({type = "cmd_response", action = "reset_stack", result = "ok"})

    elseif cmd_data.action == "reboot_mcu" then
        log.info("CMD", "重启模块")
        pm.reboot()
        send_to_uart({type = "cmd_response", action = "reboot_mcu", result = "ok"})
    else
        send_to_uart({type = "error", msg = "unknown command"})
    end
end

-- =================================================================================
-- 异步任务消费者 (短信发送)
-- =================================================================================

sys.taskInit(function()
    while true do
        if #sms_send_queue > 0 then
            -- 取出第一个任务 (FIFO)
            local task = table.remove(sms_send_queue, 1)

            log.info("Queue", "开始处理短信任务", task.to, "剩余任务:", #sms_send_queue)

            -- 执行发送（阻塞操作，独占硬件）
            local result = sms.sendLong(task.to, task.content).wait(CONFIG.SMS_SEND_TIMEOUT)
            local success = (result == true)

            if not success then
                log.warn("Queue", "短信发送失败或超时", task.to)
            end

            -- 发送结果回传给上位机
            send_to_uart({
                type = "sms_send_result",
                success = success,
                request_id = task.request_id,
                to = task.to,
                timestamp = os.time()
            })

            -- 适当延时，给模组喘息时间
            sys.wait(500)
        else
            -- 队列为空，等待新消息通知或超时轮询
            sys.waitUntil("NEW_SMS_TASK", 1000)
        end
    end
end)

-- =================================================================================
-- 事件监听区
-- =================================================================================

sys.subscribe("SMS_INC", function(phone, content)
    log.info("Event", "收到短信:", phone)
    local msg = {
        type = "incoming_sms",
        timestamp = os.time(),
        from = phone,
        content = content
    }
    table.insert(msg_buffer, msg)
    if #msg_buffer > CONFIG.MAX_BUFFER_SIZE then
        table.remove(msg_buffer, 1) -- 移除旧的
    end
    sys.publish("NEW_MSG_IN_BUFFER")
end)

sys.subscribe("SIM_IND", function(status)
    send_to_uart({type = "sim_event", status = status})
end)

-- 来电事件处理
sys.subscribe("CC_IND", function(state)
    if state == "READY" then
        log.info("Call", "通话准备完成")

    elseif state == "INCOMINGCALL" then
        -- 有电话呼入
        if call_ring_count == 0 then
            log.info("Call", "检测到来电")
            local phone_num = cc.lastNum()
            log.info("Call", "来电号码:", phone_num or "unknown")

            -- 转发来电通知到 UART
            send_to_uart({
                type = "incoming_call",
                timestamp = os.time(),
                from = phone_num or "unknown"
            })
        end

        call_ring_count = call_ring_count + 1

    elseif state == "DISCONNECTED" then
        -- 电话被挂断
        log.info("Call", "通话结束")
        call_ring_count = 0
        send_to_uart({
            type = "call_disconnected",
            timestamp = os.time()
        })
    end
end)

-- =================================================================================
-- 串口接收处理 (通用处理函数)
-- =================================================================================

local function handle_uart_receive(id, len)
    local chunk = uart.read(id, len)
    if not chunk or #chunk == 0 then return end

    -- 获取当前串口的缓冲区
    local buffer_table = uart_buffers[id]
    table.insert(buffer_table, chunk)

    -- 合并检查完整包
    -- 策略：为了性能，这里每次都会 concat，如果数据量特别大建议优化策略
    -- 但考虑到指令一般很短，直接 concat 问题不大
    local current_str = table.concat(buffer_table)

    while true do
        local start_pos = current_str:find("CMD_START:", 1, true)
        if not start_pos then
            -- 如果缓冲区过大且无头，清空
            if #current_str > CONFIG.UART_BUFFER_MAX then
                log.error("UART", "Buffer Overflow - 清空缓冲区")
                uart_buffers[id] = {}
                send_to_uart({type="error", msg="Buffer overflow, cleared"})
            end
            break
        end

        local end_pos = current_str:find(":CMD_END", start_pos + 10, true)
        if not end_pos then break end  -- 数据不完整，等待

        -- 提取 JSON 部分
        local json_str = current_str:sub(start_pos + 10, end_pos - 1)

        -- 移除已处理部分
        current_str = current_str:sub(end_pos + 8)

        -- 更新缓冲区 table
        uart_buffers[id] = {}
        if #current_str > 0 then
            table.insert(uart_buffers[id], current_str)
        end

        if #json_str > 0 then
            local success, cmd = pcall(json.decode, json_str)
            if success and cmd then
                process_uart_command(cmd)
            else
                log.warn("UART", "JSON解析失败:", json_str:sub(1, 50))
                send_to_uart({type="error", msg="Invalid JSON"})
            end
        end

        -- 继续循环处理可能存在的下一个包
    end
end

-- 注册两个串口的监听事件
uart.on(uart_usb, "receive", handle_uart_receive)
uart.on(uart_phy, "receive", handle_uart_receive)

-- =================================================================================
-- 后台维护任务
-- =================================================================================

sys.taskInit(function()
    while true do
        if #msg_buffer == 0 then
            sys.waitUntil("NEW_MSG_IN_BUFFER")
        end
        while #msg_buffer > 0 do
            local msg = table.remove(msg_buffer, 1)
            if msg then
                send_to_uart(msg)
                sys.wait(50)
            end
        end
        if collectgarbage("count") > 1024 then
            collectgarbage("collect")
        end
    end
end)

sys.taskInit(function()
    sys.wait(5000)
    local info = get_mobile_info()
    send_to_uart({
        type = "system_ready",
        project = PROJECT,
        version = VERSION,
        imei = info.imei,
        data_disabled = true
    })
    while true do
        sys.wait(CONFIG.HEARTBEAT_INTERVAL)
        info = get_mobile_info()
        send_to_uart({
            type = "heartbeat",
            imei = info.imei,
            rssi = info.rssi,
            signal_level = info.signal_level,
            signal_desc = info.signal_desc,
            net_reg = info.is_registered,
            flymode = info.flymode,
            sim_ready = info.sim_ready,
            mem = math.floor(collectgarbage("count"))
        })
    end
end)

sys.run()

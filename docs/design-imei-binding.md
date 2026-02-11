# 基于 IMEI 的设备持久化绑定方案设计

## 1. 问题背景

在 Linux 系统（如树莓派、OpenWrt）中，USB 转串口设备的路径（`/dev/ttyUSB0`）是动态分配的。
- **端口漂移 (Port Drifting)**: 设备重启或重新插拔后，`/dev/ttyUSB0` 可能变成 `/dev/ttyUSB1`。
- **多设备混淆**: 当连接多个相同型号的 4G 模组时，无法区分哪个是“移动卡”，哪个是“联通卡”。

## 2. 核心思路

将设备的**唯一身份标识**从“易变的路径”转移到“不变的硬件 ID”。

- **主键 (Identity)**: **IMEI** (国际移动设备识别码)。存储在 4G 模组内部，绝对唯一，不受 USB 接口或操作系统影响。
- **辅助键 (Cache)**: **USB 硬件指纹** (`VID:PID:SerialNumber`)。用于快速过滤和初筛。
- **易变键 (Path)**: **串口路径** (`/dev/ttyUSB0`)。仅作为最后一次连接成功的缓存，失效时自动更新。

## 3. 数据库模型变更

修改 `internal/models/device.go`，在 `devices` 表中增加关键字段：

```go
type Device struct {
    ID          string `gorm:"primaryKey" json:"id"`
    Name        string `json:"name"` // 用户定义的名称，如 "移动卡"

    // 【核心身份】
    Imei        string `gorm:"unique;index" json:"imei"` // 4G模组的 IMEI，绝对唯一

    // 【辅助身份】
    HardwareID  string `gorm:"index" json:"hardwareId"` // USB 指纹 (VID:PID:SN)

    // 【易变路径】
    SerialPort  string `json:"serialPort"` // 最后一次成功的路径，如 /dev/ttyUSB0

    // ... 其他字段 (Status, Enabled, etc.)
}
```

## 4. 启动流程与自愈机制 (Auto-Repair)

在 `DeviceManager.startDevice` 中实现**双阶段绑定**：

### 阶段一：快速路径 (Fast Path)
1. 读取数据库中记录的 `SerialPort` 和 `HardwareID`。
2. 检查当前系统该路径是否存在，且 USB 指纹是否匹配。
3. 如果匹配，尝试打开串口并发送 `get_status` 指令。
4. **验证 IMEI**: 如果返回的 IMEI 与数据库一致 -> **启动成功**。
5. 任何一步失败（路径不存在、指纹不匹配、IMEI 不对），进入阶段二。

### 阶段二：慢速扫描 (Full Scan / Self-Healing)
1. 遍历系统当前所有**未被占用**的 USB 串口（`/dev/ttyUSB*`）。
2. **逐个探测**:
   - 打开串口。
   - 发送探测指令 (获取 IMEI)。
   - 等待响应 (超时 1-2秒)。
   - 关闭串口。
3. **匹配逻辑**:
   - 找到 IMEI 匹配的端口 -> **更新数据库** (`SerialPort` 和 `HardwareID`) -> **启动成功**。
   - 遍历结束仍未找到 -> **启动失败** (标记为 `OFFLINE`，记录日志 "Device missing")。

## 5. 代码实现细节

### Lua 脚本端 (`main.lua`)
现有的 `get_status` 命令已经返回了 IMEI，无需大改，只需确保在系统启动完成前也能响应基本的查询。

```lua
-- 现有的处理逻辑（已满足要求）
elseif cmd_data.action == "get_status" then
    send_to_uart({
        type = "status_response",
        imei = mobile.imei(), -- 关键字段
        -- ...
    })
```

### Go 后端 (`DeviceManager`)

```go
// 伪代码示例
func (dm *DeviceManager) startDevice(device *models.Device) error {
    // 1. Fast Path
    if dm.verifyDevice(device.SerialPort, device.Imei) {
        return dm.launchSerialService(device.SerialPort, device)
    }

    // 2. Slow Path (Port Scanning)
    availablePorts := dm.scanSystemPorts()
    for _, port := range availablePorts {
        detectedImei := dm.probePort(port)
        if detectedImei == device.Imei {
            // 找到设备，自动修复数据库
            dm.updateDevicePath(device.ID, port)
            return dm.launchSerialService(port, device)
        }
    }

    return fmt.Errorf("设备未找到 (IMEI: %s)", device.Imei)
}
```

## 6. 用户体验优化

1. **添加设备向导**:
   - 用户点击“扫描设备”。
   - 后端扫描所有串口，列出：`Port: /dev/ttyUSB0 | IMEI: 8669xxxx | Name: Air780E`。
   - 用户选择一个，系统自动记录 IMEI。
2. **状态指示**:
   - 如果发生端口漂移自动修复，前端显示通知："检测到设备端口变更，已自动重连"。

## 7. 优势

- **零配置维护**: 用户即使随意插拔 USB 线，系统也能自动找到对应的设备。
- **高可靠性**: 彻底解决 Linux 串口乱序问题。
- **兼容性**: 不依赖 Linux 特定的 udev 规则，适用于 Docker 容器（需映射 `/dev`）。

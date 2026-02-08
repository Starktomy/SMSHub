# SMSHub

> 🔌 多设备短信网关 | Multi-Device SMS Gateway

基于 Air780 LTE 模块的自托管短信网关系统，支持多设备集中管理，并实时转发到钉钉、企业微信、飞书、Telegram、邮件等渠道。

[项目说明](https://www.typesafe.cn/posts/air780e-giffgaff/)

## ✨ 功能特性

### 📱 短信管理
- 短信实时收发
- 短信记录与搜索
- 来电通知转发
- 验证码自动识别（规划中）

### 🖥️ 多设备管理
- 支持多个 Air780 设备同时连接
- 设备状态实时监控（在线/离线/信号强度）
- 串口自动发现
- 设备分组管理

### 🔔 通知渠道
- 钉钉机器人
- 企业微信机器人
- 飞书机器人
- Telegram Bot
- 邮件通知
- 自定义 Webhook

### ⏰ 定时任务
- 计划任务发送短信
- 指定设备发送
- 执行状态追踪

## 📸 截图

![screenshot1.png](screenshots/screenshot1.png)
![screenshot2.png](screenshots/screenshot2.png)

## 🛠️ 已测试设备

| 设备型号 | 状态 | 备注 |
|---------|------|------|
| Air780EHV | ✅ 推荐 | |
| Air780EHM | ✅ 推荐 | |
| Air780E | ⚠️ 可用 | 过时设备，不建议购买 |
| Air780EPV | ⚠️ 可用 | 过时设备，不建议购买 |

## 🚀 快速开始

### 1. 硬件准备

- 准备 1-20 个 Air780 模块
- 插入有效的 SIM 卡
- 通过 USB 连接到服务器

### 2. 烧录 Lua 脚本

使用 [**LuaTools**](https://docs.openluat.com/air780epm/common/Luatools/) 烧录 `main.lua` 脚本。

> 第一次烧录需要点击「下载底层和脚本」

![write.png](screenshots/write.png)

### 3. 测试串口通信

```bash
# 查看可用串口
ls -la /dev/ttyUSB* /dev/ttyACM*

# 测试命令（使用 screen 或 minicom）
screen /dev/ttyUSB0 115200

# 发送测试命令
CMD_START:{"action":"get_status"}:CMD_END
```

![test.png](screenshots/test.png)

### 4. 部署服务

#### Docker 方式（推荐）

```bash
# 创建目录
mkdir -p /opt/smshub && cd /opt/smshub

# 下载配置文件
wget https://raw.githubusercontent.com/Starktomy/SMSHub/main/docker-compose.yml
wget https://raw.githubusercontent.com/Starktomy/SMSHub/main/config.example.yaml -O config.yaml

# 修改配置（映射 USB 路径、设置密码等）
vim docker-compose.yml
vim config.yaml

# 启动服务
docker-compose up -d
```

#### 原生方式

```bash
# 下载
wget https://github.com/Starktomy/SMSHub/releases/latest/download/smshub-linux-amd64.tar.gz

# 解压
tar -zxvf smshub-linux-amd64.tar.gz -C /opt/
mv /opt/smshub-linux-amd64 /opt/smshub

# 创建数据目录
mkdir -p /opt/smshub/data

# 创建系统服务
cat <<EOF > /etc/systemd/system/smshub.service
[Unit]
Description=SMSHub Service
After=network.target

[Service]
User=root
WorkingDirectory=/opt/smshub
ExecStart=/opt/smshub/smshub
TimeoutSec=0
RestartSec=10
Restart=always
LimitNOFILE=1048576

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
systemctl daemon-reload
systemctl enable smshub
systemctl start smshub
```

### 5. 访问 Web 界面

打开浏览器访问 `http://your-server:8080`

## 📡 API 接口

### 设备管理

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/api/devices` | 获取设备列表 |
| POST | `/api/devices` | 添加设备 |
| PUT | `/api/devices/:id` | 更新设备 |
| DELETE | `/api/devices/:id` | 删除设备 |
| POST | `/api/devices/:id/enable` | 启用设备 |
| POST | `/api/devices/:id/disable` | 禁用设备 |
| POST | `/api/devices/:id/flymode` | 设置飞行模式 |
| POST | `/api/devices/:id/reboot` | 重启设备 |
| GET | `/api/devices/discover` | 扫描可用串口 |

### 短信发送

| 方法 | 路径 | 说明 |
|------|------|------|
| POST | `/api/devices/:id/sms` | 指定设备发送 |
| POST | `/api/sms/send` | 自动选择设备发送 |
| POST | `/api/sms/batch` | 多收件人发送 |

**多收件人发送示例：**

```bash
curl -X POST http://localhost:8080/api/sms/batch \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "recipients": ["+8613800138000", "+8613900139000"],
    "content": "测试短信",
    "strategy": "round_robin"
  }'
```

## ⚙️ 配置说明

参考 [config.example.yaml](config.example.yaml) 文件：

```yaml
app:
  jwt:
    secret: "your-secret-key"
    expiresHours: 168
  users:
    admin: "$2a$10$..."  # bcrypt 加密的密码
```

## 🏗️ 技术栈

- **后端**: Go + Echo + GORM + SQLite
- **前端**: React + TypeScript + TailwindCSS + Shadcn/UI
- **设备端**: Lua (LuaT 平台)

## 📝 更新日志

### v1.1.0
- ✨ 新增多设备管理功能
- ✨ 新增设备分组管理
- 🔧 优化 Lua 脚本，添加超时机制
- 🔧 优化心跳检测机制

### v1.0.x
- 基础短信收发功能
- 通知渠道支持
- 定时任务功能

## 📄 License

MIT License

## 🙏 致谢

- [LuaT](https://www.openluat.com/) - Air780 开发平台
- [dushixiang](https://github.com/dushixiang) - 原项目作者

---

## ⚠️ 免责声明

本项目仅供学习和测试用途。使用本软件时，请务必：

1. **遵守当地法律法规** - 未经授权发送短信可能违反相关法律
2. **遵守运营商规则** - 请遵循 SIM 卡所属运营商的使用条款和服务协议
3. **禁止用于非法用途** - 包括但不限于垃圾短信、诈骗、骚扰等

作者不对因使用本软件而产生的任何法律责任或损失承担责任。使用本软件即表示您已理解并同意以上条款。

# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 交流规范

- 在所有对话中称呼用户为「老板」。
- 使用中文进行交流。

## 项目概述

SMSHub 是一个基于 Air780 LTE 模块的自托管多设备短信网关系统。通过 UART 串口接收短信和来电，存储到 SQLite 数据库，并转发通知到钉钉、企业微信、飞书、Telegram、邮件或自定义 Webhook。提供 Web 管理界面和 HTTP API。

## 构建与运行命令

```bash
# 开发
make run                # go run 直接运行后端
make dev                # 构建开发二进制文件（不含前端、不压缩）-> bin/smshub

# 前端（需要 yarn）
make build-web          # cd web && yarn && yarn build
cd web && yarn dev      # Vite 开发服务器（HMR 热更新）

# 前端代码检查
cd web && yarn lint      # ESLint

# 发布构建（前端 + 后端，UPX 压缩）
make build              # 等同于 make build-release

# 跨平台构建
make build-servers      # 8 个平台（linux/windows/darwin/freebsd）
make build-linux        # linux amd64+arm64（用于 Docker 镜像）

make clean              # 清理 bin/ 和 web/dist/
```

## 测试

```bash
# 运行所有后端测试
go test ./internal/service/... ./internal/handler/... -v -count=1

# 运行单个测试
go test ./internal/service/ -run TestOIDCStateStoreConcurrency -v

# 运行特定包的测试
go test ./internal/handler/... -v
```

测试文件位于对应包目录下（`*_test.go`），覆盖 service 层和 handler 层。

## 代码修改工作流

修改代码后，必须按以下流程完成：

1. **编写/更新测试** — 为所有修改的文件编写对应的测试用例
2. **格式化与检查** — 运行以下命令确保代码风格和规范：
   - 后端：`go fmt ./... && go vet ./...`
   - 前端：`cd web && npm run lint`
3. **运行测试** — `go test ./internal/service/... ./internal/handler/... -v -count=1`，确保全部通过
4. **本地编译** — `CGO_ENABLED=0 go build -o bin/smshub cmd/serv/main.go`，确保编译成功
5. **Commit & Push** — 只有上述步骤都通过后，才提交并推送到远程仓库

## 架构

### 三层系统

1. **设备固件** (`main.lua`) — 运行在 Air780 LTE 模块上的 Lua 脚本。处理短信/来电事件，每 60 秒发送心跳，通过双 UART（115200 波特率）处理指令。

2. **Go 后端** (`cmd/serv/main.go` → `internal/`) — Echo v4 HTTP 服务器，内嵌前端 SPA。使用纯 Go 实现的 GORM SQLite（无 CGO），`CGO_ENABLED=0` 静态编译。

3. **React 前端** (`web/`) — Vite + React 19 + TypeScript + Tailwind CSS v4 + Shadcn/UI + React Query + React Router v7。

### UART 串口协议

Go 服务器与 Air780 模块之间的文本帧协议：
- **服务器 → 设备**: `CMD_START:{json}:CMD_END\r\n`
- **设备 → 服务器**: `SMS_START:{json}:SMS_END\r\n`

协议解析：`internal/service/serial_protocol.go`
按 `type` 字段路由消息：`internal/service/serial_router.go`

### 后端分层架构

```
handler/    → HTTP 处理器（Echo）
service/    → 业务逻辑层
repo/       → 数据库访问层（GORM）
models/     → 数据模型
config/     → 配置结构体（Viper/YAML）
```

所有依赖注入和初始化在 `internal/app.go` 中完成 — 这是引导入口，负责初始化数据库、服务、路由和后台协程。

### 核心服务

- **`DeviceManager`** (`device_manager.go`) — 多设备编排器。每个设备对应一个 `SerialService`。短信发送策略：`auto`（最佳信号）、`round_robin`（轮询）、`random`（随机）、`signal_best`（信号最强）。
- **`SerialService`** (`serial_service.go`) — 单串口管理器。指数退避重连（5s–60s）。支持串口自动检测，读取并解析 UART 帧。
- **`Notifier`** (`notifier.go`) — 多渠道通知分发器。支持钉钉（HMAC 签名）、企业微信、飞书（签名）、Telegram（代理支持）、SMTP 邮件、自定义 Webhook（fasttemplate 变量：`{{from}}`、`{{content}}`、`{{type}}`、`{{timestamp}}`）。
- **`SchedulerService`** — 基于 `robfig/cron/v3` 的定时任务。

### 前端 API 层

`web/src/api/` 下的类型化 API 客户端模块（auth、devices、messages、serial、property、scheduled_task）。使用 React Query（`@tanstack/react-query`）管理状态。

## 配置

复制 `config.example.yaml` 为 `config.yaml`。关键配置项：
- 数据库：SQLite 路径（`./data/app.db`）
- 服务器：监听地址（`:8080`）
- JWT 密钥（留空自动生成）、过期时间（默认 168 小时）
- 串口（留空 = 自动检测）
- OIDC 单点登录（可选）

## Go 模块

模块路径：`github.com/Starktomy/smshub`，Go 1.25+。版本号通过 `-ldflags` 在构建时从 git tag 注入。

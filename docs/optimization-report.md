# SMSHub 项目优化分析报告

## 一、Go 后端（30 个问题）

### 🔴 P0 — 必须立即修复

**1. OIDC stateStore 并发不安全**
`internal/service/oidc_service.go` — `stateStore` 是普通 map，多个 goroutine 同时读写会导致 panic。需要加 `sync.RWMutex` 保护。

**2. Notifier 类型断言无检查**
`internal/service/notifier.go:168-173, 262` — 直接 `config["apiToken"].(string)` 无 `ok` 检查，配置缺失时整个服务 panic。

**3. TextMessageHandler.Delete 缺少 return**
`internal/handler/text_message_handler.go:33-35` — `id == ""` 的 if 分支没有 `return`，空 ID 会继续执行删除逻辑，属于 Bug。

**4. JWT 密钥管理不足**
`internal/app.go:161-163` — 空密钥时用随机 UUID，重启后所有 token 失效。应强制配置或持久化生成的密钥。

### 🟠 P1 — 高优先级

**5. 错误消息泄露内部信息**
多个 handler 直接把 `err.Error()` 返回给前端，可能暴露数据库路径、SQL 错误等敏感信息。

**6. DeviceManager 锁持有时间过长**
`internal/service/device_manager.go:135-177` — `startDevice` 整个函数 defer 解锁，创建串口服务期间阻塞其他设备操作。

**7. HTTP Client 未复用**
`internal/service/notifier.go:319, 356` — 每次通知请求都 `new(http.Client)`，应使用全局 Client 复用连接池。

**8. 没有优雅关闭机制**
`cmd/serv/main.go` — 无信号处理，无 graceful shutdown。进程被杀时数据库连接、串口、正在处理的请求全部被中断。

### 🟡 P2 — 中优先级

**9. 会话列表全量加载 + 冒泡排序**
`internal/service/text_message_service.go:129-185` — 将所有消息加载到内存后手动 O(n²) 排序，应改用 SQL `GROUP BY + MAX` 聚合查询。

**10. 会话和消息列表缺少分页**
`internal/handler/text_message_handler.go:80-89` — 数据量大时返回全量数据会超时。

**11. 缓存过期时间不一致**
`internal/service/property_service.go` — 初始化用 1 分钟，Set 时用 1 小时，逻辑混乱。

**12. goroutine 泄漏风险**
`internal/app.go:144` 和 `device_manager.go:170` — 启动的 goroutine 无 context 控制，应用退出时无法清理。

**13. 缺少审计日志**
清空消息、删除设备等危险操作没有记录操作人信息。

---

## 二、React 前端（27 个问题）

### 🔴 P0 — 必须立即修复

**1. Token 存储在 localStorage**
`src/pages/Login.tsx:44`、`src/api/client.ts:39` — 容易被 XSS 攻击窃取，应改用 httpOnly cookie。

**2. Layout 绕过统一 API Client**
`src/components/Layout.tsx:37-44` — 直接用 `fetch` 请求 `/api/devices`，绕过了 `client.ts` 的统一认证和 401 处理逻辑。

**3. API 错误处理不一致**
`src/api/client.ts:66-74` — 假设错误响应都是文本，但后端返回的是 JSON，导致错误信息解析失败。

**4. 大量 `any` 类型**
`NotificationChannels.tsx`、`ScheduledTasksConfig.tsx`、`Devices.tsx` 等多处使用 `any`，丧失类型安全。

### 🟠 P1 — 高优先级

**5. NotificationChannels 单文件超 1087 行**
应拆分为 DingtalkConfig、WecomConfig、FeishuConfig、TelegramConfig、EmailConfig、WebhookConfig 等子组件。

**6. 缺少 Error Boundary**
组件崩溃会导致整个应用白屏，需要在 `App.tsx` 外层包裹 Error Boundary。

**7. Dashboard 混用 useState/useEffect 和 React Query**
`src/pages/Dashboard.tsx:13-31` — stats 数据手动 setInterval 轮询，应统一用 React Query 的 `refetchInterval`。

**8. 缺少 404 路由**
`src/App.tsx` 没有 catch-all 路由，访问不存在页面显示空白。

### 🟡 P2 — 中优先级

**9. 轮询间隔不统一且过于频繁**
Layout(5s)、Dashboard(10s)、Messages(5s)、SerialControl(10s) — 多页面同时轮询给服务器造成压力，应考虑 WebSocket 或拉长间隔。

**10. 消息列表缺少虚拟滚动**
`src/pages/Messages.tsx:437-489` — 消息量大时 `.map()` 全量渲染性能差，需 `react-window`。

**11. 大量重复代码**
加载动画、空状态、时间格式化、状态颜色映射在多个页面重复实现，应抽取公共组件和 utils。

**12. 缺少代码分割**
所有页面打包在同一个 bundle 中，应使用 `React.lazy` 实现路由级懒加载。

**13. useEffect 依赖项缺失**
`src/pages/NotificationChannels.tsx:174` — 依赖数组不完整，存在 stale closure 风险。

---

## 三、Lua 固件（main.lua）

### 🔴 高优先级

**1. 短信发送无重试机制** — 发送失败直接丢弃，应加入有限次重试。

**2. 串口初始化错误处理不足** — 初始化失败时只打印日志，不会重试或报警。

**3. 看门狗超时时间过短** — 9 秒喂一次（3 秒间隔），串口阻塞时可能误触发重启。

### 🟡 中优先级

**4. 消息队列无容量上限** — 大量待发短信可能撑爆内存。

**5. 缺少内存管理** — 没有定期 `collectgarbage()` 调用。

---

## 四、基础设施与 CI/CD

### 🔴 高优先级

**1. CI/CD 中 Go 和 Node 版本不存在**
`.github/workflows/*.yml` — `go-version: '1.25'` 和 `node-version: '24'` 都不存在，CI 必定失败。应改为 `'1.22'` 和 `'22'`。

**2. Dockerfile 使用 `alpine:latest`**
不可复现构建，应固定为 `alpine:3.19`。

**3. Docker 容器以 root 运行**
`Dockerfile` 没有创建非 root 用户，存在安全风险。

**4. docker-compose 无资源限制和健康检查**
容器可能无限占用资源，且无法自动检测服务异常。

### 🟡 中优先级

**5. Release 缺少 Checksum 和签名** — 用户无法验证下载文件完整性。

**6. 缺少镜像安全扫描** — 应添加 Trivy 扫描。

**7. config.example.yaml 日志级别为 debug** — 生产模板不应默认 debug 级别。

---

## 五、优化优先级总览

| 优先级 | 类别 | 数量 | 建议 |
|--------|------|------|------|
| **P0** | 安全漏洞 + Bug | **11** | 立即修复 |
| **P1** | 并发安全 + 架构 | **12** | 本轮迭代完成 |
| **P2** | 性能 + 可维护性 | **18** | 排入后续计划 |
| **P3** | 文档 + 体验优化 | **~10** | 按需推进 |

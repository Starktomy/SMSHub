# forwarder-sms 功能移植计划

> 本文档记录从 `forwarder-sms` 项目中值得移植到 `uart_sms_forwarder` 的功能特性。

## 项目对比

| 特性 | forwarder-sms | uart_sms_forwarder |
|------|---------------|-------------------|
| **平台** | Cloudflare Workers (无服务器) | Go + SQLite (本地部署) |
| **短信来源** | HTTP API 接收 | 直接连接 Air780 硬件 |
| **存储** | Cloudflare KV | SQLite |
| **通知渠道** | Bark, 飞书, 企业微信, 钉钉 | 钉钉, 企业微信, 飞书, Webhook, 邮件, Telegram |

---

## 待移植功能

### 1. 验证码自动提取 ⭐⭐⭐

**优先级**: 🔴 高
**工作量**: 低
**价值**: 高

**描述**: 自动从短信内容中提取验证码，便于用户快速查看和复制。

**forwarder-sms 实现参考** (`src/utils/validator.js`):
```javascript
const CODE_PATTERNS = [
    /(?:验证码|校验码|确认码|动态码|安全码|code)[是为：:\s]*(\d{4,8})/i,
    /(\d{4,8})(?:\s*(?:是|为)?(?:您的)?(?:验证码|校验码|确认码|动态码|安全码))/i,
    /\b(\d{6})\b/,  // 标准6位验证码
    /\b(\d{4})\b/,  // 标准4位验证码
];

function extractCode(text) {
    for (const pattern of CODE_PATTERNS) {
        const match = text.match(pattern);
        if (match && match[1]) {
            return match[1];
        }
    }
    return null;
}
```

**移植方案**:
1. 在 `internal/service/` 新建 `code_extractor.go`
2. 实现正则匹配逻辑
3. 修改 `TextMessage` 模型，添加 `VerificationCode` 字段
4. 短信保存时自动提取验证码
5. 通知消息中高亮显示验证码
6. 前端 Dashboard 显示最近验证码列表

**Go 实现示例**:
```go
package service

import "regexp"

var codePatterns = []*regexp.Regexp{
    regexp.MustCompile(`(?i)(?:验证码|校验码|确认码|动态码|安全码|code)[是为：:\s]*(\d{4,8})`),
    regexp.MustCompile(`(\d{4,8})(?:\s*(?:是|为)?(?:您的)?(?:验证码|校验码|确认码|动态码|安全码))`),
    regexp.MustCompile(`\b(\d{6})\b`),
    regexp.MustCompile(`\b(\d{4})\b`),
}

func ExtractVerificationCode(content string) string {
    for _, pattern := range codePatterns {
        if matches := pattern.FindStringSubmatch(content); len(matches) > 1 {
            return matches[1]
        }
    }
    return ""
}
```

---

### 2. Bark 推送支持 ⭐⭐⭐

**优先级**: 🔴 高
**工作量**: 中
**价值**: 高 (iOS 用户首选)

**描述**: Bark 是 iOS 平台最流行的推送工具，支持自建服务器。

**forwarder-sms 实现参考** (`src/utils/bark.js`):
```javascript
async function sendBarkNotification(env, title, body) {
    const deviceKeys = env.BARK_DEVICE_KEYS?.split(',') || [];
    const serverUrl = env.BARK_SERVER_URL || 'https://api.day.app';

    const results = await Promise.all(deviceKeys.map(async (key) => {
        const url = `${serverUrl}/${key.trim()}/${encodeURIComponent(title)}/${encodeURIComponent(body)}`;
        const response = await fetch(url);
        return response.ok;
    }));

    return results.some(r => r);
}
```

**移植方案**:
1. 在 `internal/service/notifier.go` 添加 `SendBark` 方法
2. 支持配置项:
   - `bark_server_url`: 服务器地址（默认 `https://api.day.app`）
   - `bark_device_keys`: 设备密钥列表（逗号分隔，支持多设备）
   - `bark_sound`: 推送声音
   - `bark_icon`: 推送图标
3. 前端添加 Bark 配置页面

**Go 实现示例**:
```go
type BarkConfig struct {
    ServerURL  string   `json:"serverUrl"`
    DeviceKeys []string `json:"deviceKeys"`
    Sound      string   `json:"sound"`
    Icon       string   `json:"icon"`
}

func (n *Notifier) SendBark(ctx context.Context, config BarkConfig, title, body string) error {
    for _, key := range config.DeviceKeys {
        url := fmt.Sprintf("%s/%s/%s/%s",
            config.ServerURL,
            key,
            url.PathEscape(title),
            url.PathEscape(body))

        resp, err := http.Get(url)
        if err != nil {
            return err
        }
        resp.Body.Close()
    }
    return nil
}
```

---

### 3. 消息去重机制 ⭐⭐

**优先级**: 🟡 中
**工作量**: 低
**价值**: 中

**描述**: 基于内容哈希的去重，防止网络抖动导致重复通知。

**forwarder-sms 实现参考** (`src/handlers/sms.js`):
```javascript
async function hashContent(content) {
    const encoder = new TextEncoder();
    const data = encoder.encode(content);
    const hashBuffer = await crypto.subtle.digest('SHA-256', data);
    const hashArray = Array.from(new Uint8Array(hashBuffer));
    return hashArray.map(b => b.toString(16).padStart(2, '0')).join('');
}

async function isDuplicate(env, deviceId, content) {
    const hash = await hashContent(`${deviceId}:${content}`);
    const key = `dedup:${hash}`;
    const existing = await env.SMS_KV.get(key);
    if (existing) return true;

    await env.SMS_KV.put(key, '1', { expirationTtl: 300 }); // 5分钟过期
    return false;
}
```

**移植方案**:
1. 在短信保存前检查是否重复
2. 使用内存缓存（go-cache）或数据库记录
3. 配置去重时间窗口（默认5分钟）

**Go 实现示例**:
```go
import (
    "crypto/sha256"
    "encoding/hex"
    "time"

    "github.com/patrickmn/go-cache"
)

var dedupCache = cache.New(5*time.Minute, 10*time.Minute)

func IsDuplicate(deviceID, content string) bool {
    hash := sha256.Sum256([]byte(deviceID + ":" + content))
    key := "dedup:" + hex.EncodeToString(hash[:])

    if _, found := dedupCache.Get(key); found {
        return true
    }

    dedupCache.Set(key, true, cache.DefaultExpiration)
    return false
}
```

---

### 4. 并行通知发送 ⭐⭐

**优先级**: 🟡 中
**工作量**: 低
**价值**: 中

**描述**: 当前通知是串行发送，可优化为并行发送提升性能。

**forwarder-sms 实现参考** (`src/handlers/sms.js`):
```javascript
const [feishuResult, wecomResult, dingtalkResult, barkResult] = await Promise.all([
    sendFeishuNotification(env, title, content, deviceId, code),
    sendWecomNotification(env, title, content, deviceId, code),
    sendDingtalkNotification(env, title, content, deviceId, code),
    sendBarkNotification(env, title, body),
]);
```

**移植方案**:
修改 `serial_handlers_sms.go` 中的 `sendNotificationMessage` 方法。

**Go 实现示例**:
```go
func (s *SerialService) sendNotificationMessage(ctx context.Context, msg NotificationMessage) {
    channels, err := s.propertyService.GetNotificationChannelConfigs(ctx)
    if err != nil {
        s.logger.Error("获取通知渠道配置失败", zap.Error(err))
        return
    }

    message := msg.String()

    var wg sync.WaitGroup
    for _, channel := range channels {
        if !channel.Enabled {
            continue
        }

        wg.Add(1)
        go func(ch NotificationChannel) {
            defer wg.Done()

            var sendErr error
            switch ch.Type {
            case "dingtalk":
                sendErr = s.notifier.SendDingTalkByConfig(ctx, ch.Config, message)
            case "wecom":
                sendErr = s.notifier.SendWeComByConfig(ctx, ch.Config, message)
            case "feishu":
                sendErr = s.notifier.SendFeishuByConfig(ctx, ch.Config, message)
            case "bark":
                sendErr = s.notifier.SendBarkByConfig(ctx, ch.Config, message)
            // ... 其他渠道
            }

            if sendErr != nil {
                s.logger.Error("发送通知失败", zap.String("type", ch.Type), zap.Error(sendErr))
            }
        }(channel)
    }

    wg.Wait()
}
```

---

### 5. 速率限制 ⭐

**优先级**: 🟢 低
**工作量**: 中
**价值**: 低 (本地部署场景不太需要)

**描述**: 滑动窗口限流，防止 API 滥用。

**forwarder-sms 实现参考** (`src/utils/rateLimit.js`):
```javascript
async function checkRateLimit(env, identifier, limit = 10, windowSecs = 60) {
    const key = `ratelimit:${identifier}`;
    const now = Date.now();
    const windowStart = now - windowSecs * 1000;

    // 获取当前窗口内的请求记录
    const data = await env.SMS_KV.get(key, 'json') || [];
    const recentRequests = data.filter(ts => ts > windowStart);

    if (recentRequests.length >= limit) {
        return { allowed: false, remaining: 0 };
    }

    recentRequests.push(now);
    await env.SMS_KV.put(key, JSON.stringify(recentRequests), { expirationTtl: windowSecs * 2 });

    return { allowed: true, remaining: limit - recentRequests.length };
}
```

**移植方案**:
- 使用 `golang.org/x/time/rate` 标准库
- 或使用 Redis 实现分布式限流
- 主要用于保护 `/api/sms/batch` 等批量接口

---

## 实施计划

### Phase 1: 高优先级功能
- [ ] 验证码自动提取
- [ ] Bark 推送支持

### Phase 2: 中优先级功能
- [ ] 消息去重机制
- [ ] 并行通知发送

### Phase 3: 低优先级功能
- [ ] 速率限制

---

## 参考资料

- forwarder-sms 源码: `/root/sms/forwarder-sms/`
- Bark 官方文档: https://github.com/Finb/Bark
- Bark API: https://api.day.app

---

*文档创建时间: 2024-02*
*最后更新: 2024-02*

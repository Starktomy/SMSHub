package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"go.uber.org/zap"
)

func newTestNotifier() *Notifier {
	logger, _ := zap.NewDevelopment()
	return NewNotifier(logger)
}

func TestNewNotifierHasHTTPClient(t *testing.T) {
	n := newTestNotifier()
	if n.httpClient == nil {
		t.Fatal("httpClient 不应为 nil")
	}
	if n.httpClient.Timeout != 10*time.Second {
		t.Errorf("httpClient.Timeout 应为 10s，实际为 %v", n.httpClient.Timeout)
	}
}

func TestSendTelegramByConfigMissingAPIToken(t *testing.T) {
	n := newTestNotifier()
	ctx := context.Background()

	// 缺少 apiToken
	config := map[string]interface{}{
		"userid": "12345",
	}
	err := n.SendTelegramByConfig(ctx, config, "test")
	if err == nil {
		t.Error("缺少 apiToken 应返回错误")
	}

	// apiToken 为空字符串
	config["apiToken"] = ""
	err = n.SendTelegramByConfig(ctx, config, "test")
	if err == nil {
		t.Error("空 apiToken 应返回错误")
	}

	// apiToken 为非字符串类型
	config["apiToken"] = 12345
	err = n.SendTelegramByConfig(ctx, config, "test")
	if err == nil {
		t.Error("非字符串 apiToken 应返回错误")
	}
}

func TestSendTelegramByConfigMissingUserID(t *testing.T) {
	n := newTestNotifier()
	ctx := context.Background()

	config := map[string]interface{}{
		"apiToken": "valid-token",
	}
	err := n.SendTelegramByConfig(ctx, config, "test")
	if err == nil {
		t.Error("缺少 userid 应返回错误")
	}
}

func TestSendCustomWebhookMissingURL(t *testing.T) {
	n := newTestNotifier()
	ctx := context.Background()

	config := map[string]interface{}{
		"body": `{"msg": "{{content}}"}`,
	}
	msg := NotificationMessage{Type: "sms", From: "123", Content: "test", Timestamp: time.Now().Unix()}
	err := n.SendWebhookByConfig(ctx, config, msg)
	if err == nil {
		t.Error("缺少 url 应返回错误")
	}
}

func TestSendCustomWebhookMissingBody(t *testing.T) {
	n := newTestNotifier()
	ctx := context.Background()

	config := map[string]interface{}{
		"url": "https://example.com/webhook",
	}
	msg := NotificationMessage{Type: "sms", From: "123", Content: "test", Timestamp: time.Now().Unix()}
	err := n.SendWebhookByConfig(ctx, config, msg)
	if err == nil {
		t.Error("缺少 body 应返回错误")
	}
}

func TestSendCustomWebhookContentTypeDefault(t *testing.T) {
	// 启动测试 HTTP 服务器
	var receivedContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := newTestNotifier()
	ctx := context.Background()

	// 不设置 contentType，应默认使用 application/json
	config := map[string]interface{}{
		"url":  server.URL,
		"body": `{"msg": "{{content}}"}`,
	}
	msg := NotificationMessage{Type: "sms", From: "123", Content: "hello", Timestamp: time.Now().Unix()}
	err := n.SendWebhookByConfig(ctx, config, msg)
	if err != nil {
		t.Fatalf("发送请求失败: %v", err)
	}
	if receivedContentType != "application/json" {
		t.Errorf("默认 Content-Type 应为 application/json，实际为 %s", receivedContentType)
	}
}

func TestSendCustomWebhookContentTypeNonString(t *testing.T) {
	var receivedContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n := newTestNotifier()
	ctx := context.Background()

	// contentType 为非字符串类型，应默认使用 application/json
	config := map[string]interface{}{
		"url":         server.URL,
		"body":        `{"msg": "test"}`,
		"contentType": 123,
	}
	msg := NotificationMessage{Type: "sms", From: "123", Content: "hello", Timestamp: time.Now().Unix()}
	err := n.SendWebhookByConfig(ctx, config, msg)
	if err != nil {
		t.Fatalf("发送请求失败: %v", err)
	}
	if receivedContentType != "application/json" {
		t.Errorf("非字符串 contentType 应回退为 application/json，实际为 %s", receivedContentType)
	}
}

func TestSendJSONRequestSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Content-Type 应为 application/json")
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	n := newTestNotifier()
	ctx := context.Background()

	body := map[string]string{"test": "value"}
	resp, err := n.sendJSONRequest(ctx, server.URL, body)
	if err != nil {
		t.Fatalf("sendJSONRequest 失败: %v", err)
	}
	if resp == nil {
		t.Fatal("响应不应为 nil")
	}
}

func TestSendJSONRequestNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error"))
	}))
	defer server.Close()

	n := newTestNotifier()
	ctx := context.Background()

	_, err := n.sendJSONRequest(ctx, server.URL, map[string]string{})
	if err == nil {
		t.Error("非 2xx 响应应返回错误")
	}
}

func TestSendDingTalkByConfigMissingSecretKey(t *testing.T) {
	n := newTestNotifier()
	ctx := context.Background()

	config := map[string]interface{}{}
	err := n.SendDingTalkByConfig(ctx, config, "test")
	if err == nil {
		t.Error("缺少 secretKey 应返回错误")
	}
}

func TestSendWeComByConfigMissingSecretKey(t *testing.T) {
	n := newTestNotifier()
	ctx := context.Background()

	config := map[string]interface{}{}
	err := n.SendWeComByConfig(ctx, config, "test")
	if err == nil {
		t.Error("缺少 secretKey 应返回错误")
	}
}

func TestSendFeishuByConfigMissingSecretKey(t *testing.T) {
	n := newTestNotifier()
	ctx := context.Background()

	config := map[string]interface{}{}
	err := n.SendFeishuByConfig(ctx, config, "test")
	if err == nil {
		t.Error("缺少 secretKey 应返回错误")
	}
}

func TestNotificationMessageString(t *testing.T) {
	msg := NotificationMessage{
		Type:      "sms",
		From:      "+8613800138000",
		Content:   "测试短信内容",
		Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC).Unix(),
	}
	str := msg.String()
	if str == "" {
		t.Error("String() 不应返回空字符串")
	}

	callMsg := NotificationMessage{
		Type:      "call",
		From:      "+8613800138000",
		Timestamp: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC).Unix(),
	}
	callStr := callMsg.String()
	if callStr == "" {
		t.Error("来电通知 String() 不应返回空字符串")
	}
}

func TestBuildProxyURL(t *testing.T) {
	u, err := buildProxyURL("http://proxy.example.com:8080", "user", "pass")
	if err != nil {
		t.Fatalf("buildProxyURL 失败: %v", err)
	}
	if u.Host != "proxy.example.com:8080" {
		t.Errorf("host 应为 proxy.example.com:8080，实际为 %s", u.Host)
	}
	if u.User == nil {
		t.Fatal("User 不应为 nil")
	}
	password, _ := u.User.Password()
	if password != "pass" {
		t.Errorf("密码应为 pass，实际为 %s", password)
	}
}

func TestBuildProxyURLNoAuth(t *testing.T) {
	u, err := buildProxyURL("http://proxy.example.com:8080", "", "")
	if err != nil {
		t.Fatalf("buildProxyURL 失败: %v", err)
	}
	if u.User != nil {
		t.Error("无认证时 User 应为 nil")
	}
}

func TestHTTPClientReusedAcrossCalls(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"ok": true}`))
	}))
	defer server.Close()

	n := newTestNotifier()
	ctx := context.Background()

	// 多次调用应使用同一个 httpClient
	clientBefore := n.httpClient
	n.sendJSONRequest(ctx, server.URL, map[string]string{})
	n.sendJSONRequest(ctx, server.URL, map[string]string{})
	clientAfter := n.httpClient

	if clientBefore != clientAfter {
		t.Error("httpClient 应在多次调用间保持不变")
	}
	if callCount != 2 {
		t.Errorf("服务器应收到 2 次请求，实际为 %d", callCount)
	}
}

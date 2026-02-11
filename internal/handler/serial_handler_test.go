package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestSerialHandlerSendSMSMissingFields(t *testing.T) {
	e := echo.New()

	// 空手机号
	body := `{"to": "", "content": "hello"}`
	req := httptest.NewRequest(http.MethodPost, "/api/serial/sms", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &SerialHandler{}
	err := h.SendSMS(c)
	if err != nil {
		t.Fatalf("SendSMS 不应返回 Echo 错误: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("空手机号应返回 400，实际为 %d", rec.Code)
	}

	// 空内容
	body2 := `{"to": "13800138000", "content": ""}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/serial/sms", strings.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	err = h.SendSMS(c2)
	if err != nil {
		t.Fatalf("SendSMS 不应返回 Echo 错误: %v", err)
	}

	if rec2.Code != http.StatusBadRequest {
		t.Errorf("空内容应返回 400，实际为 %d", rec2.Code)
	}
}

func TestSerialHandlerSendSMSInvalidJSON(t *testing.T) {
	e := echo.New()

	body := `{invalid}`
	req := httptest.NewRequest(http.MethodPost, "/api/serial/sms", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &SerialHandler{}
	err := h.SendSMS(c)
	if err != nil {
		t.Fatalf("SendSMS 不应返回 Echo 错误: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("无效 JSON 应返回 400，实际为 %d", rec.Code)
	}
}

func TestSerialHandlerSetFlymodeInvalidJSON(t *testing.T) {
	e := echo.New()

	body := `{invalid}`
	req := httptest.NewRequest(http.MethodPost, "/api/serial/flymode", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &SerialHandler{}
	err := h.SetFlymode(c)
	if err != nil {
		t.Fatalf("SetFlymode 不应返回 Echo 错误: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("无效 JSON 应返回 400，实际为 %d", rec.Code)
	}
}

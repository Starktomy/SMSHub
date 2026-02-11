package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestDeviceHandlerCreateMissingSerialPort(t *testing.T) {
	e := echo.New()

	body := `{"name": "test", "serialPort": "", "enabled": true}`
	req := httptest.NewRequest(http.MethodPost, "/api/devices", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &DeviceHandler{}
	err := h.Create(c)
	if err != nil {
		t.Fatalf("Create 不应返回 Echo 错误: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("空串口路径应返回 400，实际为 %d", rec.Code)
	}
}

func TestDeviceHandlerCreateInvalidJSON(t *testing.T) {
	e := echo.New()

	body := `{invalid json}`
	req := httptest.NewRequest(http.MethodPost, "/api/devices", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &DeviceHandler{}
	err := h.Create(c)
	if err != nil {
		t.Fatalf("Create 不应返回 Echo 错误: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("无效 JSON 应返回 400，实际为 %d", rec.Code)
	}
}

func TestDeviceHandlerUpdateInvalidJSON(t *testing.T) {
	e := echo.New()

	body := `{invalid}`
	req := httptest.NewRequest(http.MethodPut, "/api/devices/test-id", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("test-id")

	h := &DeviceHandler{}
	err := h.Update(c)
	if err != nil {
		t.Fatalf("Update 不应返回 Echo 错误: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("无效 JSON 应返回 400，实际为 %d", rec.Code)
	}
}

func TestDeviceHandlerSendSMSMissingFields(t *testing.T) {
	e := echo.New()

	// 测试空手机号
	body := `{"to": "", "content": "hello"}`
	req := httptest.NewRequest(http.MethodPost, "/api/devices/test-id/sms", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("test-id")

	h := &DeviceHandler{}
	err := h.SendSMS(c)
	if err != nil {
		t.Fatalf("SendSMS 不应返回 Echo 错误: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("空手机号应返回 400，实际为 %d", rec.Code)
	}

	// 测试空内容
	body2 := `{"to": "13800138000", "content": ""}`
	req2 := httptest.NewRequest(http.MethodPost, "/api/devices/test-id/sms", strings.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)
	c2.SetParamNames("id")
	c2.SetParamValues("test-id")

	err = h.SendSMS(c2)
	if err != nil {
		t.Fatalf("SendSMS 不应返回 Echo 错误: %v", err)
	}

	if rec2.Code != http.StatusBadRequest {
		t.Errorf("空内容应返回 400，实际为 %d", rec2.Code)
	}
}

func TestDeviceHandlerAutoSendSMSMissingFields(t *testing.T) {
	e := echo.New()

	body := `{"to": "", "content": "hello"}`
	req := httptest.NewRequest(http.MethodPost, "/api/sms/send", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &DeviceHandler{}
	err := h.AutoSendSMS(c)
	if err != nil {
		t.Fatalf("AutoSendSMS 不应返回 Echo 错误: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("空手机号应返回 400，实际为 %d", rec.Code)
	}
}

func TestDeviceHandlerBatchSendSMSMissingRecipients(t *testing.T) {
	e := echo.New()

	body := `{"recipients": [], "content": "hello"}`
	req := httptest.NewRequest(http.MethodPost, "/api/sms/batch", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	h := &DeviceHandler{}
	err := h.BatchSendSMS(c)
	if err != nil {
		t.Fatalf("BatchSendSMS 不应返回 Echo 错误: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("空收件人列表应返回 400，实际为 %d", rec.Code)
	}
}

func TestDeviceHandlerSetFlymodeInvalidJSON(t *testing.T) {
	e := echo.New()

	body := `{invalid}`
	req := httptest.NewRequest(http.MethodPost, "/api/devices/test-id/flymode", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("test-id")

	h := &DeviceHandler{}
	err := h.SetFlymode(c)
	if err != nil {
		t.Fatalf("SetFlymode 不应返回 Echo 错误: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("无效 JSON 应返回 400，实际为 %d", rec.Code)
	}
}

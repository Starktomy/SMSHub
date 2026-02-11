package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
)

func TestTextMessageHandlerDeleteEmptyID(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodDelete, "/api/messages/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues("")

	h := &TextMessageHandler{}
	err := h.Delete(c)
	if err != nil {
		t.Fatalf("Delete 不应返回 Echo 错误: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("空 ID 应返回 400，实际为 %d", rec.Code)
	}
}

func TestTextMessageHandlerGetConversationMessagesEmptyPeer(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/api/messages/conversations//messages", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("peer")
	c.SetParamValues("")

	h := &TextMessageHandler{}
	err := h.GetConversationMessages(c)
	if err != nil {
		t.Fatalf("GetConversationMessages 不应返回 Echo 错误: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("空 peer 应返回 400，实际为 %d", rec.Code)
	}
}

func TestTextMessageHandlerDeleteConversationEmptyPeer(t *testing.T) {
	e := echo.New()

	req := httptest.NewRequest(http.MethodDelete, "/api/messages/conversations/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("peer")
	c.SetParamValues("")

	h := &TextMessageHandler{}
	err := h.DeleteConversation(c)
	if err != nil {
		t.Fatalf("DeleteConversation 不应返回 Echo 错误: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Errorf("空 peer 应返回 400，实际为 %d", rec.Code)
	}
}

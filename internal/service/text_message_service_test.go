package service

import (
	"context"
	"testing"
	"time"

	"github.com/Starktomy/smshub/internal/models"
	"github.com/Starktomy/smshub/internal/repo"
	"go.uber.org/zap"
)

func TestTextMessageService_Conversations(t *testing.T) {
	db := setupTestDB(t)
	msgRepo := repo.NewTextMessageRepo(db)
	logger := zap.NewExample()
	svc := NewTextMessageService(logger, msgRepo)
	ctx := context.Background()

	// 准备测试数据
	messages := []models.TextMessage{
		// 与 10086 的会话 (Incoming)
		{
			ID:        "msg1",
			From:      "10086",
			To:        "Self",
			Content:   "Bill info",
			Type:      models.MessageTypeIncoming,
			Status:    models.MessageStatusReceived,
			CreatedAt: time.Now().Add(-1 * time.Hour).UnixMilli(),
		},
		{
			ID:        "msg2",
			From:      "10086",
			To:        "Self",
			Content:   "New bill",
			Type:      models.MessageTypeIncoming,
			Status:    models.MessageStatusReceived,
			CreatedAt: time.Now().UnixMilli(),
		},
		// 与 12345 的会话 (Outgoing)
		{
			ID:        "msg3",
			From:      "Self",
			To:        "12345",
			Content:   "Hello",
			Type:      models.MessageTypeOutgoing,
			Status:    models.MessageStatusSent,
			CreatedAt: time.Now().Add(-2 * time.Hour).UnixMilli(),
		},
	}

	for _, msg := range messages {
		if err := svc.Save(ctx, &msg); err != nil {
			t.Fatalf("Failed to save message: %v", err)
		}
	}

	// 测试 GetConversations
	t.Run("GetConversations", func(t *testing.T) {
		convs, err := svc.GetConversations(ctx)
		if err != nil {
			t.Fatalf("GetConversations failed: %v", err)
		}

		if len(convs) != 2 {
			t.Errorf("Expected 2 conversations, got %d", len(convs))
		}

		// 检查 10086 会话
		var found10086 bool
		for _, c := range convs {
			if c.Peer == "10086" {
				found10086 = true
				if c.MessageCount != 2 {
					t.Errorf("Expected 2 messages for 10086, got %d", c.MessageCount)
				}
				if c.LastMessage.Content != "New bill" {
					t.Errorf("Expected last message 'New bill', got '%s'", c.LastMessage.Content)
				}
			}
		}
		if !found10086 {
			t.Error("Conversation with 10086 not found")
		}
	})

	// 测试 GetConversationMessages
	t.Run("GetConversationMessages", func(t *testing.T) {
		msgs, err := svc.GetConversationMessages(ctx, "10086")
		if err != nil {
			t.Fatalf("GetConversationMessages failed: %v", err)
		}
		if len(msgs) != 2 {
			t.Errorf("Expected 2 messages, got %d", len(msgs))
		}
	})

	// 测试 DeleteConversation
	t.Run("DeleteConversation", func(t *testing.T) {
		err := svc.DeleteConversation(ctx, "10086")
		if err != nil {
			t.Fatalf("DeleteConversation failed: %v", err)
		}

		// 验证删除
		msgs, err := svc.GetConversationMessages(ctx, "10086")
		if err != nil {
			t.Fatalf("GetConversationMessages after delete failed: %v", err)
		}
		if len(msgs) != 0 {
			t.Errorf("Expected 0 messages after delete, got %d", len(msgs))
		}
	})
}

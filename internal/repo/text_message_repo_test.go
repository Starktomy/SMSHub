package repo

import (
	"context"
	"testing"
	"time"

	"github.com/Starktomy/smshub/internal/models"
)

func TestTextMessageRepoCRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTextMessageRepo(db)
	ctx := context.Background()

	// 1. Create
	msg := &models.TextMessage{
		ID:        "msg-1",
		From:      "+8613800138000",
		Content:   "Hello World",
		Type:      models.MessageTypeIncoming,
		CreatedAt: time.Now().UnixMilli(),
	}

	if err := repo.Create(ctx, msg); err != nil {
		t.Fatalf("创建短信失败: %v", err)
	}

	// 2. FindAll (简单验证)
	msgs, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("查找所有短信失败: %v", err)
	}
	if len(msgs) != 1 {
		t.Errorf("短信数量期望 1，实际 %d", len(msgs))
	}

	// 3. Delete
	if err := repo.DeleteById(ctx, "msg-1"); err != nil {
		t.Fatalf("删除短信失败: %v", err)
	}

	msgs, _ = repo.FindAll(ctx)
	if len(msgs) != 0 {
		t.Error("删除后短信数量应为 0")
	}
}

func TestTextMessageRepoStats(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTextMessageRepo(db)
	ctx := context.Background()

	now := time.Now()
	yesterday := now.Add(-24 * time.Hour)

	// 今日短信
	repo.Create(ctx, &models.TextMessage{ID: "1", CreatedAt: now.UnixMilli(), Content: "Today"})
	repo.Create(ctx, &models.TextMessage{ID: "2", CreatedAt: now.UnixMilli(), Content: "Today 2"})

	// 昨日短信
	repo.Create(ctx, &models.TextMessage{ID: "3", CreatedAt: yesterday.UnixMilli(), Content: "Yesterday"})

	count, err := repo.CountToday(ctx)
	if err != nil {
		t.Fatalf("统计今日短信失败: %v", err)
	}

	if count != 2 {
		t.Errorf("今日短信数量期望 2，实际 %d", count)
	}

	total, err := repo.CountAll(ctx)
	if err != nil {
		t.Fatalf("统计总短信失败: %v", err)
	}
	if total != 3 {
		t.Errorf("总短信数量期望 3，实际 %d", total)
	}
}

func TestTextMessageRepoConversations(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTextMessageRepo(db)
	ctx := context.Background()

	// 模拟会话数据
	// 会话 A: +8613800000001 (2条消息)
	if err := repo.Create(ctx, &models.TextMessage{
		ID: "1", From: "+8613800000001", To: "Self", Type: models.MessageTypeIncoming,
		CreatedAt: 1000, Content: "Msg 1",
	}); err != nil {
		t.Fatalf("创建消息1失败: %v", err)
	}
	if err := repo.Create(ctx, &models.TextMessage{
		ID: "2", From: "Self", To: "+8613800000001", Type: models.MessageTypeOutgoing,
		CreatedAt: 2000, Content: "Reply 1",
	}); err != nil {
		t.Fatalf("创建消息2失败: %v", err)
	}

	// 会话 B: +8613800000002 (1条消息)
	if err := repo.Create(ctx, &models.TextMessage{
		ID: "3", From: "+8613800000002", To: "Self", Type: models.MessageTypeIncoming,
		CreatedAt: 3000, Content: "Msg 2",
	}); err != nil {
		t.Fatalf("创建消息3失败: %v", err)
	}

	// 测试 GetConversations
	// 注意：由于我们在 TextMessageService 中重写了 GetConversations 逻辑，Repo 层的方法可能表现不同
	// 这里我们主要测试 Repo 层的基础查询能力

	// 测试 FindByPeer
	msgs, err := repo.FindByPeer(ctx, "+8613800000001")
	if err != nil {
		t.Fatalf("FindByPeer 失败: %v", err)
	}
	if len(msgs) != 2 {
		// Debug info
		var all []models.TextMessage
		db.Find(&all)
		t.Logf("Database content: %+v", all)
		t.Errorf("会话 A 消息数量期望 2，实际 %d", len(msgs))
	}
}

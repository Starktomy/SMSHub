package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/Starktomy/smshub/internal/models"
	"github.com/Starktomy/smshub/internal/repo"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TextMessageService 短信服务
type TextMessageService struct {
	repo   *repo.TextMessageRepo
	logger *zap.Logger
}

// NewTextMessageService 创建短信服务实例
func NewTextMessageService(logger *zap.Logger, repo *repo.TextMessageRepo) *TextMessageService {
	return &TextMessageService{
		repo:   repo,
		logger: logger,
	}
}

// Stats 统计信息
type Stats struct {
	TotalCount    int64 `json:"totalCount"`
	IncomingCount int64 `json:"incomingCount"`
	OutgoingCount int64 `json:"outgoingCount"`
	TodayCount    int64 `json:"todayCount"`
}

// Conversation 会话信息
type Conversation struct {
	Peer         string              `json:"peer"`         // 对方号码
	LastMessage  *models.TextMessage `json:"lastMessage"`  // 最后一条消息
	MessageCount int64               `json:"messageCount"` // 消息总数
	UnreadCount  int64               `json:"unreadCount"`  // 未读数量（暂时为0）
}

// Save 保存短信记录
func (s *TextMessageService) Save(ctx context.Context, msg *models.TextMessage) error {
	if err := s.repo.Save(ctx, msg); err != nil {
		s.logger.Error("保存短信记录失败", zap.Error(err), zap.String("id", msg.ID))
		return fmt.Errorf("保存短信记录失败: %w", err)
	}
	return nil
}

// Get 获取单条短信记录
func (s *TextMessageService) Get(ctx context.Context, id string) (*models.TextMessage, error) {
	msg, err := s.repo.FindById(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("短信记录不存在")
		}
		s.logger.Error("获取短信记录失败", zap.Error(err), zap.String("id", id))
		return nil, fmt.Errorf("获取短信记录失败: %w", err)
	}
	return &msg, nil
}

// Delete 删除单条短信记录
func (s *TextMessageService) Delete(ctx context.Context, id string) error {
	if err := s.repo.DeleteById(ctx, id); err != nil {
		s.logger.Error("删除短信记录失败", zap.Error(err), zap.String("id", id))
		return fmt.Errorf("删除短信记录失败: %w", err)
	}
	s.logger.Info("删除短信记录成功", zap.String("id", id))
	return nil
}

// Clear 清空所有短信记录
func (s *TextMessageService) Clear(ctx context.Context) error {
	db := s.repo.GetDB(ctx)
	if err := db.Where("1 = 1").Delete(&models.TextMessage{}).Error; err != nil {
		s.logger.Error("清空短信记录失败", zap.Error(err))
		return fmt.Errorf("清空短信记录失败: %w", err)
	}
	s.logger.Info("清空短信记录成功")
	return nil
}

// GetStats 获取统计信息
func (s *TextMessageService) GetStats(ctx context.Context) (*Stats, error) {
	db := s.repo.GetDB(ctx)

	stats := &Stats{}

	// 总数
	if err := db.Model(&models.TextMessage{}).Count(&stats.TotalCount).Error; err != nil {
		return nil, fmt.Errorf("统计总数失败: %w", err)
	}

	// 接收数量
	if err := db.Model(&models.TextMessage{}).Where("type = ?", "incoming").Count(&stats.IncomingCount).Error; err != nil {
		return nil, fmt.Errorf("统计接收数量失败: %w", err)
	}

	// 发送数量
	if err := db.Model(&models.TextMessage{}).Where("type = ?", "outgoing").Count(&stats.OutgoingCount).Error; err != nil {
		return nil, fmt.Errorf("统计发送数量失败: %w", err)
	}

	// 今日数量（按 created_at 字段）
	todayStart := time.Now().Truncate(24 * time.Hour).UnixMilli()
	if err := db.Model(&models.TextMessage{}).Where("created_at >= ?", todayStart).Count(&stats.TodayCount).Error; err != nil {
		return nil, fmt.Errorf("统计今日数量失败: %w", err)
	}

	return stats, nil
}

func (s *TextMessageService) UpdateStatusById(ctx context.Context, id string, status models.MessageStatus) error {
	return s.repo.UpdateColumnsById(ctx, id, map[string]interface{}{
		"status": status,
	})
}

// GetConversations 获取会话列表（按对方号码分组）
func (s *TextMessageService) GetConversations(ctx context.Context) ([]*Conversation, error) {
	db := s.repo.GetDB(ctx)

	// 使用 SQL 聚合查询获取每个 peer 的消息计数和最后消息时间
	type peerSummary struct {
		Peer         string
		MessageCount int64
		LastTime     int64
	}

	var incomingSummaries []peerSummary
	if err := db.Model(&models.TextMessage{}).
		Select(`from_number as peer, COUNT(*) as message_count, MAX(created_at) as last_time`).
		Where("type = ?", models.MessageTypeIncoming).
		Group(`from_number`).
		Having(`from_number != ''`).
		Find(&incomingSummaries).Error; err != nil {
		return nil, fmt.Errorf("获取接收消息统计失败: %w", err)
	}

	var outgoingSummaries []peerSummary
	if err := db.Model(&models.TextMessage{}).
		Select(`to_number as peer, COUNT(*) as message_count, MAX(created_at) as last_time`).
		Where("type = ?", models.MessageTypeOutgoing).
		Group(`to_number`).
		Having(`to_number != ''`).
		Find(&outgoingSummaries).Error; err != nil {
		return nil, fmt.Errorf("获取发送消息统计失败: %w", err)
	}

	// 合并统计结果
	peerMap := make(map[string]*Conversation)
	for _, s := range incomingSummaries {
		peerMap[s.Peer] = &Conversation{
			Peer:         s.Peer,
			MessageCount: s.MessageCount,
		}
	}
	for _, s := range outgoingSummaries {
		if conv, exists := peerMap[s.Peer]; exists {
			conv.MessageCount += s.MessageCount
		} else {
			peerMap[s.Peer] = &Conversation{
				Peer:         s.Peer,
				MessageCount: s.MessageCount,
			}
		}
	}

	// 获取每个 peer 的最后一条消息
	for peer, conv := range peerMap {
		var lastMsg models.TextMessage
		if err := db.Where("(type = ? AND from_number = ?) OR (type = ? AND to_number = ?)",
			models.MessageTypeIncoming, peer,
			models.MessageTypeOutgoing, peer,
		).Order("created_at DESC").First(&lastMsg).Error; err != nil {
			continue
		}
		conv.LastMessage = &lastMsg
		conv.UnreadCount = 0 // 暂时不实现未读计数
		peerMap[peer] = conv
	}

	// 转换为切片
	conversations := make([]*Conversation, 0, len(peerMap))
	for _, conv := range peerMap {
		if conv.LastMessage != nil {
			conversations = append(conversations, conv)
		}
	}

	// 按最后消息时间倒序排序
	sort.Slice(conversations, func(i, j int) bool {
		return conversations[i].LastMessage.CreatedAt > conversations[j].LastMessage.CreatedAt
	})

	return conversations, nil
}

// GetConversationMessages 获取指定会话的所有消息
func (s *TextMessageService) GetConversationMessages(ctx context.Context, peer string) ([]models.TextMessage, error) {
	db := s.repo.GetDB(ctx)

	var messages []models.TextMessage

	// 查询条件：(type=incoming AND from_number=peer) OR (type=outgoing AND to_number=peer)
	if err := db.Where("(type = ? AND from_number = ?) OR (type = ? AND to_number = ?)",
		models.MessageTypeIncoming, peer,
		models.MessageTypeOutgoing, peer,
	).Order("created_at ASC").Find(&messages).Error; err != nil {
		s.logger.Error("获取会话消息失败", zap.Error(err), zap.String("peer", peer))
		return nil, fmt.Errorf("获取会话消息失败: %w", err)
	}

	return messages, nil
}

// DeleteConversation 删除整个会话（与某个联系人的所有消息）
func (s *TextMessageService) DeleteConversation(ctx context.Context, peer string) error {
	db := s.repo.GetDB(ctx)

	// 删除条件：(type=incoming AND from_number=peer) OR (type=outgoing AND to_number=peer)
	result := db.Where("(type = ? AND from_number = ?) OR (type = ? AND to_number = ?)",
		models.MessageTypeIncoming, peer,
		models.MessageTypeOutgoing, peer,
	).Delete(&models.TextMessage{})

	if result.Error != nil {
		s.logger.Error("删除会话失败", zap.Error(result.Error), zap.String("peer", peer))
		return fmt.Errorf("删除会话失败: %w", result.Error)
	}

	s.logger.Info("删除会话成功", zap.String("peer", peer), zap.Int64("deleted_count", result.RowsAffected))
	return nil
}

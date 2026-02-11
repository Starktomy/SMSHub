package repo

import (
	"context"
	"time"

	"github.com/Starktomy/smshub/internal/models"
	"github.com/go-orz/orz"
	"gorm.io/gorm"
)

func NewTextMessageRepo(db *gorm.DB) *TextMessageRepo {
	return &TextMessageRepo{
		db:         db,
		Repository: orz.NewRepository[models.TextMessage, string](db),
	}
}

type TextMessageRepo struct {
	orz.Repository[models.TextMessage, string]
	db *gorm.DB
}

func (r *TextMessageRepo) CountToday(ctx context.Context) (int64, error) {
	var count int64
	// 获取今天0点的时间戳
	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).UnixMilli()

	err := r.db.WithContext(ctx).Model(&models.TextMessage{}).
		Where("created_at >= ?", startOfDay).
		Count(&count).Error
	return count, err
}

func (r *TextMessageRepo) CountAll(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.TextMessage{}).Count(&count).Error
	return count, err
}

func (r *TextMessageRepo) FindByPeer(ctx context.Context, peer string) ([]*models.TextMessage, error) {
	var msgs []*models.TextMessage
	err := r.db.WithContext(ctx).
		Where("from_number = ? OR to_number = ?", peer, peer).
		Order("created_at desc").
		Find(&msgs).Error
	return msgs, err
}

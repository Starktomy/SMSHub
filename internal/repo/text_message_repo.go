package repo

import (
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

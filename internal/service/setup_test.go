package service

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/Starktomy/smshub/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// setupTestDB 创建内存数据库用于测试
func setupTestDB(t *testing.T) *gorm.DB {
	// 使用随机名称确保隔离
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	dbName := fmt.Sprintf("memdb_svc_%d", rng.Int())

	// 使用内存数据库
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", dbName)), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				LogLevel: logger.Error, // 仅显示错误
			},
		),
	})
	if err != nil {
		t.Fatalf("连接数据库失败: %v", err)
	}

	// 自动迁移
	err = db.AutoMigrate(
		&models.Device{},
		&models.TextMessage{},
		&models.Property{},
		&models.ScheduledTask{},
	)
	if err != nil {
		t.Fatalf("数据库迁移失败: %v", err)
	}

	return db
}

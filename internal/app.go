package internal

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/Starktomy/smshub/config"
	"github.com/Starktomy/smshub/internal/handler"
	"github.com/Starktomy/smshub/internal/middleware"
	"github.com/Starktomy/smshub/internal/models"
	"github.com/Starktomy/smshub/internal/repo"
	"github.com/Starktomy/smshub/internal/service"
	"github.com/Starktomy/smshub/internal/version"
	"github.com/Starktomy/smshub/web"
	"github.com/go-orz/orz"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Handlers 所有Handler的集合
type Handlers struct {
	Auth          *handler.AuthHandler
	Property      *handler.PropertyHandler
	TextMessage   *handler.TextMessageHandler
	Serial        *handler.SerialHandler
	ScheduledTask *handler.ScheduledTaskHandler
	Device        *handler.DeviceHandler
}

func Run(configPath string) {
	err := orz.Quick(configPath, setup)
	if err != nil {
		log.Fatal(err)
	}
}

func setup(app *orz.App) error {
	logger := app.Logger()
	db := app.GetDatabase()

	// 1. 数据库迁移
	if err := autoMigrate(db); err != nil {
		logger.Error("数据库迁移失败", zap.Error(err))
		return err
	}

	// 2. 读取应用配置
	var appConfig config.AppConfig
	_config := app.GetConfig()
	if _config != nil {
		if err := _config.App.Unmarshal(&appConfig); err != nil {
			logger.Error("读取配置失败", zap.Error(err))
			return err
		}
	}

	// 3. 设置默认值
	setDefaultConfig(&appConfig, logger)

	// 4. 初始化 Repository
	textMessageRepo := repo.NewTextMessageRepo(db)
	deviceRepo := repo.NewDeviceRepo(db)

	// 5. 初始化 Service
	propertyService := service.NewPropertyService(logger, db)
	notifier := service.NewNotifier(logger)
	textMessageService := service.NewTextMessageService(logger, textMessageRepo)

	// 初始化默认配置
	ctx := context.Background()
	if err := propertyService.InitializeDefaultConfigs(ctx); err != nil {
		logger.Error("初始化默认配置失败", zap.Error(err))
	}

	// 6. 初始化设备管理器
	deviceManager := service.NewDeviceManager(
		logger,
		deviceRepo,
		textMessageService,
		notifier,
		propertyService,
	)

	// 7. 初始化串口服务（兼容单设备模式）
	serialService := service.NewSerialService(
		logger,
		appConfig.Serial,
		textMessageService,
		notifier,
		propertyService,
	)

	// 8. 初始化定时任务服务
	schedulerService := service.NewSchedulerService(
		logger,
		db,
		serialService,
		deviceManager,
	)
	serialService.SetScheduledTaskStatusUpdater(schedulerService.UpdateLastRunStatusByMsgId)
	deviceManager.SetScheduledTaskStatusUpdater(schedulerService.UpdateLastRunStatusByMsgId)

	// 9. 初始化 OIDC 和 Account Service
	oidcService := service.NewOIDCService(logger, &appConfig)
	accountService := service.NewAccountService(logger, oidcService, &appConfig)

	// 10. 初始化 Handler
	authHandler := handler.NewAuthHandler(logger, accountService)
	propertyHandler := handler.NewPropertyHandler(logger, propertyService, notifier)
	textMessageHandler := handler.NewTextMessageHandler(logger, textMessageService, textMessageRepo)
	serialHandler := handler.NewSerialHandler(logger, serialService)
	scheduledTaskHandler := handler.NewScheduledTaskHandler(logger, schedulerService)
	deviceHandler := handler.NewDeviceHandler(logger, deviceManager)

	handlers := &Handlers{
		Auth:          authHandler,
		Property:      propertyHandler,
		TextMessage:   textMessageHandler,
		Serial:        serialHandler,
		ScheduledTask: scheduledTaskHandler,
		Device:        deviceHandler,
	}

	// 11. 设置 API 路由
	setupApi(app, handlers, &appConfig, logger)

	// 12. 启动后台服务
	background := context.Background()

	// 启动设备管理器
	if err := deviceManager.Start(background); err != nil {
		logger.Error("启动设备管理器失败", zap.Error(err))
	} else {
		logger.Info("设备管理器启动成功")
	}

	// 启动串口服务（单设备兼容模式，如果配置了串口）
	if appConfig.Serial.Port != "" {
		go serialService.Start()
	}

	// 启动定时任务服务
	if err := schedulerService.Start(background); err != nil {
		logger.Error("启动定时任务服务失败", zap.Error(err))
	} else {
		logger.Info("定时任务服务启动成功")
	}

	// 13. 注册优雅关闭钩子
	e := app.GetEcho()
	e.Server.RegisterOnShutdown(func() {
		logger.Info("开始优雅关闭...")

		// 停止定时任务
		schedulerService.Stop()

		// 停止串口服务（单设备模式）
		if appConfig.Serial.Port != "" {
			serialService.Stop()
		}

		// 停止设备管理器（内部会停止所有设备的串口服务）
		deviceManager.Stop()

		logger.Info("优雅关闭完成")
	})

	logger.Info("应用启动完成")
	return nil
}

// setDefaultConfig 设置默认配置
func setDefaultConfig(appConfig *config.AppConfig, logger *zap.Logger) {
	// JWT 默认值
	if appConfig.JWT.Secret == "" {
		appConfig.JWT.Secret = uuid.NewString()
		logger.Warn("未配置JWT密钥，使用随机UUID")
	}
	if appConfig.JWT.ExpiresHours == 0 {
		appConfig.JWT.ExpiresHours = 168 // 7天
	}
}

// autoMigrate 数据库迁移
func autoMigrate(db *gorm.DB) error {
	if err := db.AutoMigrate(
		&models.Property{},
		&models.TextMessage{},
		&models.ScheduledTask{},
		&models.Device{},
	); err != nil {
		return err
	}

	// 显式添加 ICCID 和 IMSI 列（如果不存在）
	// 使用原生 SQL 避免 GORM 列名映射问题
	columns := []struct {
		Name string
		Type string
	}{
		{"iccid", "TEXT"},
		{"imsi", "TEXT"},
		{"flymode", "BOOLEAN DEFAULT 0"},
	}
	for _, col := range columns {
		var count int64
		db.Raw("SELECT COUNT(*) FROM pragma_table_info('devices') WHERE name = ?", col.Name).Scan(&count)
		if count == 0 {
			if err := db.Exec("ALTER TABLE devices ADD COLUMN " + col.Name + " " + col.Type).Error; err != nil {
				// 忽略 "duplicate column name" 错误
				if !strings.Contains(err.Error(), "duplicate column name") {
					return err
				}
			}
		}
	}

	// 数据迁移：将旧的 from/to 字段数据迁移到新字段 from_number/to_number
	// 忽略错误，因为如果旧列不存在（全新安装），SQL 会执行失败
	_ = db.Exec("UPDATE text_messages SET from_number = `from` WHERE (from_number IS NULL OR from_number = '') AND `from` IS NOT NULL").Error
	_ = db.Exec("UPDATE text_messages SET to_number = `to` WHERE (to_number IS NULL OR to_number = '') AND `to` IS NOT NULL").Error

	return nil
}

// setupApi 设置API路由
func setupApi(app *orz.App, handlers *Handlers, appConfig *config.AppConfig, logger *zap.Logger) {
	e := app.GetEcho()

	e.Use(echomiddleware.StaticWithConfig(echomiddleware.StaticConfig{
		Skipper: func(c echo.Context) bool {
			// 不处理接口
			if strings.HasPrefix(c.Request().RequestURI, "/api") {
				return true
			}
			if strings.HasPrefix(c.Request().RequestURI, "/health") {
				return true
			}
			return false
		},
		Index:      "index.html",
		HTML5:      true,
		Browse:     false,
		IgnoreBase: false,
		Filesystem: http.FS(web.Assets()),
	}))

	// 登录路由（不需要认证）
	e.POST("/api/login", handlers.Auth.Login)
	e.GET("/api/auth/config", handlers.Auth.GetAuthConfig)
	e.GET("/api/auth/oidc/url", handlers.Auth.GetOIDCAuthURL)
	e.POST("/api/auth/oidc/callback", handlers.Auth.OIDCCallback)

	// API 路由组（需要认证）
	api := e.Group("/api")
	api.Use(middleware.JWTMiddleware(appConfig.JWT.Secret, logger))

	// Version
	api.GET("/version", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{
			"version": version.GetVersion(),
		})
	})

	// Property API
	api.GET("/properties/:id", handlers.Property.GetProperty)
	api.PUT("/properties/:id", handlers.Property.SetProperty)
	api.POST("/notifications/:type/test", handlers.Property.TestNotificationChannel)

	// TextMessage API
	api.GET("/messages/stats", handlers.TextMessage.GetStats)
	api.GET("/messages/conversations", handlers.TextMessage.GetConversations)
	api.GET("/messages/conversations/:peer/messages", handlers.TextMessage.GetConversationMessages)
	api.DELETE("/messages/conversations/:peer", handlers.TextMessage.DeleteConversation)
	api.DELETE("/messages/:id", handlers.TextMessage.Delete)
	api.DELETE("/messages", handlers.TextMessage.Clear)

	// Serial API
	api.POST("/serial/sms", handlers.Serial.SendSMS)
	api.GET("/serial/status", handlers.Serial.GetStatus) // 包含移动网络信息
	api.POST("/serial/flymode", handlers.Serial.SetFlymode)
	api.POST("/serial/reboot", handlers.Serial.RebootMcu)

	// ScheduledTask API (RESTful)
	api.GET("/scheduled-tasks", handlers.ScheduledTask.List)
	api.GET("/scheduled-tasks/:id", handlers.ScheduledTask.Get)
	api.POST("/scheduled-tasks", handlers.ScheduledTask.Create)
	api.PUT("/scheduled-tasks/:id", handlers.ScheduledTask.Update)
	api.DELETE("/scheduled-tasks/:id", handlers.ScheduledTask.Delete)
	api.POST("/scheduled-tasks/:id/trigger", handlers.ScheduledTask.Trigger)

	// Device API
	api.GET("/devices", handlers.Device.List)
	api.GET("/devices/discover", handlers.Device.Discover)
	api.GET("/devices/groups", handlers.Device.GetGroups)
	api.GET("/devices/stats", handlers.Device.GetStats)
	api.POST("/devices", handlers.Device.Create)
	api.GET("/devices/:id", handlers.Device.Get)
	api.PUT("/devices/:id", handlers.Device.Update)
	api.DELETE("/devices/:id", handlers.Device.Delete)
	api.POST("/devices/:id/enable", handlers.Device.Enable)
	api.POST("/devices/:id/disable", handlers.Device.Disable)
	api.POST("/devices/:id/flymode", handlers.Device.SetFlymode)
	api.POST("/devices/:id/reboot", handlers.Device.Reboot)
	api.GET("/devices/:id/status", handlers.Device.GetStatus)
	api.POST("/devices/:id/sms", handlers.Device.SendSMS)

	// SMS API (enhanced)
	api.POST("/sms/send", handlers.Device.AutoSendSMS)
	api.POST("/sms/batch", handlers.Device.BatchSendSMS)

	// 健康检查接口（无需认证）
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "ok",
		})
	})
}

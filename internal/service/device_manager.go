package service

import (
	"context"
	"fmt"
	"math/rand"
	"path/filepath"
	"sync"
	"time"

	"github.com/Starktomy/smshub/internal/models"
	"github.com/Starktomy/smshub/internal/repo"
	"github.com/google/uuid"
	"go.bug.st/serial"
	"go.uber.org/zap"
)

// SendStrategy 发送策略
type SendStrategy string

const (
	StrategyAuto       SendStrategy = "auto"        // 自动（优先信号最强）
	StrategyRoundRobin SendStrategy = "round_robin" // 轮询
	StrategyRandom     SendStrategy = "random"      // 随机
	StrategySignalBest SendStrategy = "signal_best" // 信号最强
)

const (
	// 心跳超时时间
	HeartbeatTimeout = 60 * time.Second
	// 健康检查间隔
	HealthCheckInterval = 10 * time.Second
)

// ManagedDevice 管理的设备（包含运行时状态）
type ManagedDevice struct {
	Device        *models.Device
	SerialService *SerialService
	mu            sync.RWMutex
}

// DeviceManager 多设备管理服务
type DeviceManager struct {
	logger          *zap.Logger
	repo            *repo.DeviceRepo
	textMsgService  *TextMessageService
	notifier        *Notifier
	propertyService *PropertyService

	devices   map[string]*ManagedDevice // deviceID -> ManagedDevice
	devicesMu sync.RWMutex

	// 轮询索引
	roundRobinIndex int
	roundRobinMu    sync.Mutex

	// 定时任务状态更新器
	scheduledTaskStatusUpdater ScheduledTaskStatusUpdater

	// 停止信号
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewDeviceManager 创建设备管理器
func NewDeviceManager(
	logger *zap.Logger,
	repo *repo.DeviceRepo,
	textMsgService *TextMessageService,
	notifier *Notifier,
	propertyService *PropertyService,
) *DeviceManager {
	return &DeviceManager{
		logger:          logger,
		repo:            repo,
		textMsgService:  textMsgService,
		notifier:        notifier,
		propertyService: propertyService,
		devices:         make(map[string]*ManagedDevice),
		stopCh:          make(chan struct{}),
	}
}

// SetScheduledTaskStatusUpdater 设置定时任务状态更新器
func (dm *DeviceManager) SetScheduledTaskStatusUpdater(updater ScheduledTaskStatusUpdater) {
	dm.scheduledTaskStatusUpdater = updater
}

// Start 启动设备管理器
func (dm *DeviceManager) Start(ctx context.Context) error {
	dm.logger.Info("启动设备管理器")

	// 加载所有启用的设备
	devices, err := dm.repo.FindAllEnabled(ctx)
	if err != nil {
		return fmt.Errorf("加载设备列表失败: %w", err)
	}

	// 启动每个设备
	for _, device := range devices {
		deviceCopy := device
		if err := dm.startDevice(&deviceCopy); err != nil {
			dm.logger.Error("启动设备失败",
				zap.String("id", device.ID),
				zap.String("name", device.Name),
				zap.Error(err))
		}
	}

	// 启动健康检查
	dm.wg.Add(1)
	go dm.healthCheckLoop()

	return nil
}

// Stop 停止设备管理器
func (dm *DeviceManager) Stop() {
	dm.logger.Info("停止设备管理器")
	close(dm.stopCh)
	dm.wg.Wait()

	// 停止所有设备
	dm.devicesMu.Lock()
	for _, md := range dm.devices {
		dm.logger.Info("停止设备", zap.String("id", md.Device.ID))
		md.SerialService.Stop()
	}
	dm.devices = make(map[string]*ManagedDevice)
	dm.devicesMu.Unlock()
}

// startDevice 启动单个设备
func (dm *DeviceManager) startDevice(device *models.Device) error {
	dm.devicesMu.RLock()
	_, exists := dm.devices[device.ID]
	dm.devicesMu.RUnlock()

	if exists {
		return fmt.Errorf("设备已在运行: %s", device.ID)
	}

	// 创建串口服务（耗时操作，不持锁）
	serialService := NewSerialServiceWithDeviceID(
		dm.logger.With(zap.String("device", device.Name)),
		device.SerialPort,
		device.ID,
		device.Name,
		dm.textMsgService,
		dm.notifier,
		dm.propertyService,
	)

	if dm.scheduledTaskStatusUpdater != nil {
		serialService.SetScheduledTaskStatusUpdater(dm.scheduledTaskStatusUpdater)
	}

	// 设置状态更新回调
	serialService.SetStatusUpdateCallback(func(status *StatusData) {
		dm.updateDeviceStatus(device.ID, status)
	})

	md := &ManagedDevice{
		Device:        device,
		SerialService: serialService,
	}

	// 仅在写入 map 时加锁
	dm.devicesMu.Lock()
	if _, exists := dm.devices[device.ID]; exists {
		dm.devicesMu.Unlock()
		return fmt.Errorf("设备已在运行: %s", device.ID)
	}
	dm.devices[device.ID] = md
	dm.devicesMu.Unlock()

	// 启动串口服务
	go serialService.Start()

	dm.logger.Info("设备已启动",
		zap.String("id", device.ID),
		zap.String("name", device.Name),
		zap.String("port", device.SerialPort))

	return nil
}

// stopDevice 停止单个设备
func (dm *DeviceManager) stopDevice(deviceID string) error {
	dm.devicesMu.Lock()
	defer dm.devicesMu.Unlock()

	md, exists := dm.devices[deviceID]
	if !exists {
		return fmt.Errorf("设备不存在: %s", deviceID)
	}

	// 停止串口服务
	md.SerialService.Stop()
	delete(dm.devices, deviceID)

	dm.logger.Info("设备已停止", zap.String("id", deviceID))
	return nil
}

// updateDeviceStatus 更新设备状态
func (dm *DeviceManager) updateDeviceStatus(deviceID string, status *StatusData) {
	ctx := context.Background()
	columns := map[string]any{
		"status":       models.DeviceStatusOnline,
		"last_seen_at": time.Now().UnixMilli(),
	}

	if status != nil {
		// 手机号：只有当心跳返回有值时才更新（不覆盖用户手动填写的号码）
		if status.Mobile.Number != "" {
			columns["phone_number"] = status.Mobile.Number
		}
		// IMSI：每次心跳都更新
		if status.Mobile.Imsi != "" {
			columns["imsi"] = status.Mobile.Imsi
		}

		// ICCID：直接提取
		if status.Mobile.Iccid != "" {
			columns["iccid"] = status.Mobile.Iccid
		}

		// 运营商：非空时才更新
		if status.Mobile.Operator != "" {
			columns["operator"] = status.Mobile.Operator
		}
		if status.Mobile.SimOperator != "" {
			columns["sim_operator"] = status.Mobile.SimOperator
		}
		if status.Mobile.Lac > 0 {
			columns["lac"] = status.Mobile.Lac
		}
		if status.Mobile.Cid > 0 {
			columns["cid"] = status.Mobile.Cid
		}

		// 业务数据更新
		columns["signal_level"] = status.Mobile.SignalLevel
		columns["flymode"] = status.Flymode
	}
	if err := dm.repo.UpdateColumnsById(ctx, deviceID, columns); err != nil {
		dm.logger.Error("更新设备状态失败", zap.String("id", deviceID), zap.Error(err))
	}
}

// healthCheckLoop 健康检查循环
func (dm *DeviceManager) healthCheckLoop() {
	defer dm.wg.Done()

	ticker := time.NewTicker(HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-dm.stopCh:
			return
		case <-ticker.C:
			dm.checkDevicesHealth()
		}
	}
}

// checkDevicesHealth 检查设备健康状态
func (dm *DeviceManager) checkDevicesHealth() {
	ctx := context.Background()
	now := time.Now()

	dm.devicesMu.RLock()
	deviceIDs := make([]string, 0, len(dm.devices))
	for id := range dm.devices {
		deviceIDs = append(deviceIDs, id)
	}
	dm.devicesMu.RUnlock()

	for _, id := range deviceIDs {
		device, err := dm.repo.FindById(ctx, id)
		if err != nil {
			continue
		}

		// 检查心跳超时
		lastSeen := time.UnixMilli(device.LastSeenAt)
		if now.Sub(lastSeen) > HeartbeatTimeout {
			if device.Status != models.DeviceStatusOffline {
				_ = dm.repo.UpdateColumnsById(ctx, id, map[string]any{
					"status":       models.DeviceStatusOffline,
					"signal_level": 0,
					"operator":     "",
				})
				dm.logger.Warn("设备心跳超时，标记为离线",
					zap.String("id", id),
					zap.String("name", device.Name))
			}
		}
	}
}

// ==================== 设备管理 API ====================

// GetAllDevices 获取所有设备
func (dm *DeviceManager) GetAllDevices(ctx context.Context) ([]models.Device, error) {
	return dm.repo.FindAll(ctx)
}

// GetDevice 获取单个设备
func (dm *DeviceManager) GetDevice(ctx context.Context, id string) (*models.Device, error) {
	return dm.repo.FindById(ctx, id)
}

// CreateDevice 创建设备
func (dm *DeviceManager) CreateDevice(ctx context.Context, device *models.Device) error {
	device.ID = uuid.NewString()
	device.Status = models.DeviceStatusOffline
	device.CreatedAt = time.Now().UnixMilli()
	device.UpdatedAt = time.Now().UnixMilli()

	if err := dm.repo.Create(ctx, device); err != nil {
		return err
	}

	// 如果设备启用，则启动它
	if device.Enabled {
		if err := dm.startDevice(device); err != nil {
			dm.logger.Error("创建后启动设备失败", zap.Error(err))
		}
	}

	return nil
}

// UpdateDevice 更新设备
func (dm *DeviceManager) UpdateDevice(ctx context.Context, device *models.Device) error {
	existing, err := dm.repo.FindById(ctx, device.ID)
	if err != nil {
		return err
	}

	// 检查是否需要重启设备
	needRestart := existing.SerialPort != device.SerialPort || existing.Enabled != device.Enabled

	existing.Name = device.Name
	existing.SerialPort = device.SerialPort
	existing.Enabled = device.Enabled
	existing.GroupName = device.GroupName
	existing.PhoneNumber = device.PhoneNumber
	existing.UpdatedAt = time.Now().UnixMilli()

	if err := dm.repo.Save(ctx, existing); err != nil {
		return err
	}

	if needRestart {
		// 停止旧设备
		_ = dm.stopDevice(device.ID)

		// 如果启用，启动新设备
		if device.Enabled {
			if err := dm.startDevice(existing); err != nil {
				dm.logger.Error("更新后启动设备失败", zap.Error(err))
			}
		}
	}

	return nil
}

// DeleteDevice 删除设备
func (dm *DeviceManager) DeleteDevice(ctx context.Context, id string) error {
	// 先停止设备
	_ = dm.stopDevice(id)

	return dm.repo.DeleteById(ctx, id)
}

// EnableDevice 启用设备
func (dm *DeviceManager) EnableDevice(ctx context.Context, id string) error {
	device, err := dm.repo.FindById(ctx, id)
	if err != nil {
		return err
	}

	if device.Enabled {
		return nil
	}

	device.Enabled = true
	if err := dm.repo.Save(ctx, device); err != nil {
		return err
	}

	return dm.startDevice(device)
}

// DisableDevice 禁用设备
func (dm *DeviceManager) DisableDevice(ctx context.Context, id string) error {
	if err := dm.repo.UpdateColumnsById(ctx, id, map[string]any{
		"enabled": false,
		"status":  models.DeviceStatusOffline,
	}); err != nil {
		return err
	}

	return dm.stopDevice(id)
}

// SetDeviceFlymode 设置设备飞行模式
func (dm *DeviceManager) SetDeviceFlymode(ctx context.Context, id string, enabled bool) error {
	dm.devicesMu.RLock()
	md, exists := dm.devices[id]
	dm.devicesMu.RUnlock()

	if !exists {
		return fmt.Errorf("设备不在线: %s", id)
	}

	return md.SerialService.SetFlymode(enabled)
}

// RebootDevice 重启设备
func (dm *DeviceManager) RebootDevice(ctx context.Context, id string) error {
	dm.devicesMu.RLock()
	md, exists := dm.devices[id]
	dm.devicesMu.RUnlock()

	if !exists {
		return fmt.Errorf("设备不在线: %s", id)
	}

	return md.SerialService.RebootMcu()
}

// GetDeviceStatus 获取设备状态
func (dm *DeviceManager) GetDeviceStatus(ctx context.Context, id string) (*StatusData, error) {
	dm.devicesMu.RLock()
	md, exists := dm.devices[id]
	dm.devicesMu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("设备不在线: %s", id)
	}

	return md.SerialService.GetStatus()
}

// DiscoverSerialPorts 扫描可用串口
func (dm *DeviceManager) DiscoverSerialPorts() ([]string, error) {
	ports, err := serial.GetPortsList()
	if err != nil {
		return nil, err
	}

	// 过滤出 ttyUSB 和 ttyACM 设备
	var filtered []string
	for _, port := range ports {
		base := filepath.Base(port)
		if len(base) >= 6 && (base[:6] == "ttyUSB" || base[:6] == "ttyACM") {
			filtered = append(filtered, port)
		}
	}

	return filtered, nil
}

// GetDeviceGroups 获取设备分组列表
func (dm *DeviceManager) GetDeviceGroups(ctx context.Context) ([]string, error) {
	return dm.repo.GetGroups(ctx)
}

// ==================== 短信发送 API ====================

// SendSMSByDevice 通过指定设备发送短信
func (dm *DeviceManager) SendSMSByDevice(deviceID, to, content string) (string, error) {
	dm.devicesMu.RLock()
	md, exists := dm.devices[deviceID]
	dm.devicesMu.RUnlock()

	if !exists {
		return "", fmt.Errorf("设备不在线: %s", deviceID)
	}

	return md.SerialService.SendSMS(to, content)
}

// SendSMS 自动选择设备发送短信
func (dm *DeviceManager) SendSMS(to, content string, strategy SendStrategy) (string, string, error) {
	device, err := dm.selectDevice(strategy)
	if err != nil {
		return "", "", err
	}

	msgID, err := dm.SendSMSByDevice(device.ID, to, content)
	return msgID, device.ID, err
}

// BatchSendRequest 批量发送请求
type BatchSendRequest struct {
	Recipients []string     `json:"recipients"`
	Content    string       `json:"content"`
	DeviceID   string       `json:"deviceId"`
	Strategy   SendStrategy `json:"strategy"`
}

// BatchSendResult 批量发送结果
type BatchSendResult struct {
	Recipient string `json:"recipient"`
	MessageID string `json:"messageId"`
	DeviceID  string `json:"deviceId"`
	Success   bool   `json:"success"`
	Error     string `json:"error,omitempty"`
}

// BatchSendSMS 批量发送短信
func (dm *DeviceManager) BatchSendSMS(req *BatchSendRequest) []BatchSendResult {
	results := make([]BatchSendResult, len(req.Recipients))

	for i, recipient := range req.Recipients {
		result := BatchSendResult{Recipient: recipient}

		var msgID, deviceID string
		var err error

		if req.DeviceID != "" {
			// 指定设备发送
			msgID, err = dm.SendSMSByDevice(req.DeviceID, recipient, req.Content)
			deviceID = req.DeviceID
		} else {
			// 按策略选择设备
			msgID, deviceID, err = dm.SendSMS(recipient, req.Content, req.Strategy)
		}

		if err != nil {
			result.Success = false
			result.Error = err.Error()
		} else {
			result.Success = true
			result.MessageID = msgID
			result.DeviceID = deviceID
		}

		results[i] = result
	}

	return results
}

// selectDevice 根据策略选择设备
func (dm *DeviceManager) selectDevice(strategy SendStrategy) (*models.Device, error) {
	ctx := context.Background()
	onlineDevices, err := dm.repo.FindAllOnline(ctx)
	if err != nil {
		return nil, err
	}

	if len(onlineDevices) == 0 {
		return nil, fmt.Errorf("没有可用的在线设备")
	}

	switch strategy {
	case StrategyRoundRobin:
		return dm.selectRoundRobin(onlineDevices), nil
	case StrategyRandom:
		return dm.selectRandom(onlineDevices), nil
	case StrategySignalBest, StrategyAuto:
		return dm.selectBestSignal(onlineDevices), nil
	default:
		return dm.selectBestSignal(onlineDevices), nil
	}
}

func (dm *DeviceManager) selectRoundRobin(devices []models.Device) *models.Device {
	dm.roundRobinMu.Lock()
	defer dm.roundRobinMu.Unlock()

	device := &devices[dm.roundRobinIndex%len(devices)]
	dm.roundRobinIndex++
	return device
}

func (dm *DeviceManager) selectRandom(devices []models.Device) *models.Device {
	return &devices[rand.Intn(len(devices))]
}

func (dm *DeviceManager) selectBestSignal(devices []models.Device) *models.Device {
	// devices 已按 signal_level DESC 排序
	return &devices[0]
}

// GetOnlineDeviceCount 获取在线设备数量
func (dm *DeviceManager) GetOnlineDeviceCount() int {
	dm.devicesMu.RLock()
	defer dm.devicesMu.RUnlock()
	return len(dm.devices)
}

// GetDeviceStats 获取设备统计信息
func (dm *DeviceManager) GetDeviceStats(ctx context.Context) (map[string]int64, error) {
	return dm.repo.CountByStatus(ctx)
}

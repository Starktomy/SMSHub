package repo

import (
	"context"

	"github.com/Starktomy/smshub/internal/models"
	"gorm.io/gorm"
)

// DeviceRepo 设备数据访问层
type DeviceRepo struct {
	db *gorm.DB
}

// NewDeviceRepo 创建设备仓储实例
func NewDeviceRepo(db *gorm.DB) *DeviceRepo {
	return &DeviceRepo{db: db}
}

// Create 创建设备
func (r *DeviceRepo) Create(ctx context.Context, device *models.Device) error {
	return r.db.WithContext(ctx).Create(device).Error
}

// Save 保存设备（更新或创建）
func (r *DeviceRepo) Save(ctx context.Context, device *models.Device) error {
	return r.db.WithContext(ctx).Save(device).Error
}

// FindById 根据ID查找设备
func (r *DeviceRepo) FindById(ctx context.Context, id string) (*models.Device, error) {
	var device models.Device
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// FindBySerialPort 根据串口路径查找设备
func (r *DeviceRepo) FindBySerialPort(ctx context.Context, serialPort string) (*models.Device, error) {
	var device models.Device
	err := r.db.WithContext(ctx).Where("serial_port = ?", serialPort).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

// FindAll 查找所有设备
func (r *DeviceRepo) FindAll(ctx context.Context) ([]models.Device, error) {
	var devices []models.Device
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&devices).Error
	return devices, err
}

// FindAllEnabled 查找所有启用的设备
func (r *DeviceRepo) FindAllEnabled(ctx context.Context) ([]models.Device, error) {
	var devices []models.Device
	err := r.db.WithContext(ctx).Where("enabled = ?", true).Order("created_at DESC").Find(&devices).Error
	return devices, err
}

// FindAllOnline 查找所有在线的设备
func (r *DeviceRepo) FindAllOnline(ctx context.Context) ([]models.Device, error) {
	var devices []models.Device
	err := r.db.WithContext(ctx).Where("enabled = ? AND status = ?", true, models.DeviceStatusOnline).Order("signal_level DESC").Find(&devices).Error
	return devices, err
}

// DeleteById 根据ID删除设备
func (r *DeviceRepo) DeleteById(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.Device{}).Error
}

// UpdateColumnsById 根据ID更新指定字段
func (r *DeviceRepo) UpdateColumnsById(ctx context.Context, id string, columns map[string]any) error {
	return r.db.WithContext(ctx).Model(&models.Device{}).Where("id = ?", id).Updates(columns).Error
}

// GetGroups 获取所有设备分组
func (r *DeviceRepo) GetGroups(ctx context.Context) ([]string, error) {
	var groups []string
	err := r.db.WithContext(ctx).Model(&models.Device{}).
		Distinct("group_name").
		Where("group_name != ''").
		Pluck("group_name", &groups).Error
	return groups, err
}

// CountByStatus 按状态统计设备数量
func (r *DeviceRepo) CountByStatus(ctx context.Context) (map[string]int64, error) {
	type result struct {
		Status string
		Count  int64
	}
	var results []result
	err := r.db.WithContext(ctx).Model(&models.Device{}).
		Select("status, count(*) as count").
		Group("status").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Status] = r.Count
	}
	return counts, nil
}

package repo

import (
	"context"
	"testing"
	"time"

	"github.com/Starktomy/smshub/internal/models"
)

func TestDeviceRepoCRUD(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDeviceRepo(db)
	ctx := context.Background()

	// 1. Create
	device := &models.Device{
		ID:          "device-1",
		Name:        "Test Device",
		SerialPort:  "/dev/ttyUSB0",
		Enabled:     true,
		Status:      models.DeviceStatusOnline,
		SignalLevel: 25,
		CreatedAt:   time.Now().UnixMilli(),
		UpdatedAt:   time.Now().UnixMilli(),
	}

	if err := repo.Create(ctx, device); err != nil {
		t.Fatalf("创建设备失败: %v", err)
	}

	// 2. Read (FindById)
	found, err := repo.FindById(ctx, "device-1")
	if err != nil {
		t.Fatalf("查找设备失败: %v", err)
	}
	if found.Name != "Test Device" {
		t.Errorf("设备名称不匹配，期望 'Test Device'，实际 '%s'", found.Name)
	}

	// 3. Update
	device.Name = "Updated Device"
	device.SignalLevel = 30
	if err := repo.Save(ctx, device); err != nil {
		t.Fatalf("更新设备失败: %v", err)
	}

	updated, _ := repo.FindById(ctx, "device-1")
	if updated.Name != "Updated Device" {
		t.Errorf("更新后名称不匹配，期望 'Updated Device'，实际 '%s'", updated.Name)
	}

	// 4. FindAll
	devices, err := repo.FindAll(ctx)
	if err != nil {
		t.Fatalf("获取所有设备失败: %v", err)
	}
	if len(devices) != 1 {
		t.Errorf("设备数量期望 1，实际 %d", len(devices))
	}

	// 5. Delete
	if err := repo.DeleteById(ctx, "device-1"); err != nil {
		t.Fatalf("删除设备失败: %v", err)
	}

	_, err = repo.FindById(ctx, "device-1")
	if err == nil {
		t.Error("删除后不应再找到设备")
	}
}

func TestDeviceRepoFindAllOnline(t *testing.T) {
	db := setupTestDB(t)
	repo := NewDeviceRepo(db)
	ctx := context.Background()

	// 插入一个在线设备
	repo.Create(ctx, &models.Device{
		ID:          "online-1",
		Status:      models.DeviceStatusOnline,
		SignalLevel: 20,
		Enabled:     true,
	})

	// 插入一个离线设备
	repo.Create(ctx, &models.Device{
		ID:          "offline-1",
		Status:      models.DeviceStatusOffline,
		SignalLevel: 0,
		Enabled:     true,
	})

	// 插入一个禁用设备
	repo.Create(ctx, &models.Device{
		ID:      "disabled-1",
		Status:  models.DeviceStatusOnline,
		Enabled: false,
	})

	onlineDevices, err := repo.FindAllOnline(ctx)
	if err != nil {
		t.Fatalf("查找在线设备失败: %v", err)
	}

	if len(onlineDevices) != 1 {
		t.Errorf("在线且启用的设备数量期望 1，实际 %d", len(onlineDevices))
	}
	if onlineDevices[0].ID != "online-1" {
		t.Errorf("查找到的设备 ID 错误: %s", onlineDevices[0].ID)
	}
}

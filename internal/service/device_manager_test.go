package service

import (
	"sync"
	"testing"
	"time"

	"github.com/Starktomy/smshub/internal/models"
)

func TestDeviceManagerSelectRoundRobin(t *testing.T) {
	dm := &DeviceManager{
		devices: make(map[string]*ManagedDevice),
	}

	devices := []models.Device{
		{ID: "d1", Name: "设备1"},
		{ID: "d2", Name: "设备2"},
		{ID: "d3", Name: "设备3"},
	}

	// 测试轮询选择
	selected1 := dm.selectRoundRobin(devices)
	selected2 := dm.selectRoundRobin(devices)
	selected3 := dm.selectRoundRobin(devices)
	selected4 := dm.selectRoundRobin(devices) // 应该回到第一个

	if selected1.ID != "d1" {
		t.Errorf("第1次应选择 d1，实际为 %s", selected1.ID)
	}
	if selected2.ID != "d2" {
		t.Errorf("第2次应选择 d2，实际为 %s", selected2.ID)
	}
	if selected3.ID != "d3" {
		t.Errorf("第3次应选择 d3，实际为 %s", selected3.ID)
	}
	if selected4.ID != "d1" {
		t.Errorf("第4次应回到 d1，实际为 %s", selected4.ID)
	}
}

func TestDeviceManagerSelectRandom(t *testing.T) {
	dm := &DeviceManager{
		devices: make(map[string]*ManagedDevice),
	}

	devices := []models.Device{
		{ID: "d1"},
		{ID: "d2"},
		{ID: "d3"},
	}

	// 调用多次确保不 panic
	for i := 0; i < 100; i++ {
		selected := dm.selectRandom(devices)
		if selected == nil {
			t.Fatal("selectRandom 不应返回 nil")
		}
	}
}

func TestDeviceManagerSelectBestSignal(t *testing.T) {
	dm := &DeviceManager{
		devices: make(map[string]*ManagedDevice),
	}

	devices := []models.Device{
		{ID: "d1", SignalLevel: 30},
		{ID: "d2", SignalLevel: 20},
		{ID: "d3", SignalLevel: 10},
	}

	// 已按信号排序，应选第一个
	selected := dm.selectBestSignal(devices)
	if selected.ID != "d1" {
		t.Errorf("应选择信号最强的 d1，实际为 %s", selected.ID)
	}
}

func TestDeviceManagerOnlineCount(t *testing.T) {
	dm := &DeviceManager{
		devices: make(map[string]*ManagedDevice),
	}

	if dm.GetOnlineDeviceCount() != 0 {
		t.Error("初始在线设备数量应为 0")
	}

	dm.devices["d1"] = &ManagedDevice{Device: &models.Device{ID: "d1"}}
	dm.devices["d2"] = &ManagedDevice{Device: &models.Device{ID: "d2"}}

	if dm.GetOnlineDeviceCount() != 2 {
		t.Errorf("在线设备数量应为 2，实际为 %d", dm.GetOnlineDeviceCount())
	}
}

func TestDeviceManagerRoundRobinConcurrency(t *testing.T) {
	dm := &DeviceManager{
		devices: make(map[string]*ManagedDevice),
	}

	devices := []models.Device{
		{ID: "d1"},
		{ID: "d2"},
		{ID: "d3"},
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dm.selectRoundRobin(devices)
		}()
	}
	wg.Wait()

	// 验证 roundRobinIndex 增长了 100 次
	dm.roundRobinMu.Lock()
	idx := dm.roundRobinIndex
	dm.roundRobinMu.Unlock()

	if idx != 100 {
		t.Errorf("roundRobinIndex 应为 100，实际为 %d", idx)
	}
}

func TestDeviceManagerDevicesMapConcurrency(t *testing.T) {
	dm := &DeviceManager{
		devices: make(map[string]*ManagedDevice),
		stopCh:  make(chan struct{}),
	}

	var wg sync.WaitGroup

	// 并发写入
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			id := "device-" + time.Now().String() + "-" + string(rune(idx))
			dm.devicesMu.Lock()
			dm.devices[id] = &ManagedDevice{
				Device: &models.Device{ID: id},
			}
			dm.devicesMu.Unlock()
		}(i)
	}

	// 并发读取
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			dm.devicesMu.RLock()
			_ = len(dm.devices)
			dm.devicesMu.RUnlock()
		}()
	}

	wg.Wait()
}

func TestBatchSendRequestStruct(t *testing.T) {
	req := BatchSendRequest{
		Recipients: []string{"+8613800138000", "+8613900139000"},
		Content:    "测试短信",
		Strategy:   StrategyAuto,
	}

	if len(req.Recipients) != 2 {
		t.Errorf("收件人数量应为 2，实际为 %d", len(req.Recipients))
	}
	if req.Strategy != StrategyAuto {
		t.Errorf("策略应为 auto，实际为 %s", req.Strategy)
	}
}

func TestSendStrategyConstants(t *testing.T) {
	if StrategyAuto != "auto" {
		t.Error("StrategyAuto 应为 auto")
	}
	if StrategyRoundRobin != "round_robin" {
		t.Error("StrategyRoundRobin 应为 round_robin")
	}
	if StrategyRandom != "random" {
		t.Error("StrategyRandom 应为 random")
	}
	if StrategySignalBest != "signal_best" {
		t.Error("StrategySignalBest 应为 signal_best")
	}
}

func TestHeartbeatConstants(t *testing.T) {
	if HeartbeatTimeout != 60*time.Second {
		t.Errorf("HeartbeatTimeout 应为 60s，实际为 %v", HeartbeatTimeout)
	}
	if HealthCheckInterval != 10*time.Second {
		t.Errorf("HealthCheckInterval 应为 10s，实际为 %v", HealthCheckInterval)
	}
}

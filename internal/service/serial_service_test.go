package service

import (
	"testing"
	"time"

	"github.com/Starktomy/smshub/config"
	"go.bug.st/serial"
	"go.uber.org/zap"
)

// mockSerialPort implements serial.Port for testing
type mockSerialPort struct {
	writeFunc func(p []byte) (n int, err error)
}

func (m *mockSerialPort) Read(p []byte) (n int, err error) {
	return 0, nil
}

func (m *mockSerialPort) Write(p []byte) (n int, err error) {
	if m.writeFunc != nil {
		return m.writeFunc(p)
	}
	return len(p), nil
}

func (m *mockSerialPort) Close() error                         { return nil }
func (m *mockSerialPort) SetMode(mode *serial.Mode) error      { return nil }
func (m *mockSerialPort) SetReadTimeout(t time.Duration) error { return nil }
func (m *mockSerialPort) ResetInputBuffer() error              { return nil }
func (m *mockSerialPort) ResetOutputBuffer() error             { return nil }
func (m *mockSerialPort) SetDTR(dtr bool) error                { return nil }
func (m *mockSerialPort) SetRTS(rts bool) error                { return nil }
func (m *mockSerialPort) GetModemStatusBits() (*serial.ModemStatusBits, error) {
	return &serial.ModemStatusBits{}, nil
}
func (m *mockSerialPort) Break(d time.Duration) error { return nil }
func (m *mockSerialPort) Drain() error                { return nil }

func TestSerialService_StateManagement(t *testing.T) {
	logger := zap.NewExample()
	cfg := config.SerialConfig{Port: "/dev/ttyUSB0"}
	svc := NewSerialService(logger, cfg, nil, nil, nil)

	// Test Initial State
	status, _ := svc.GetStatus()
	if status.Connected {
		t.Error("Expected initially disconnected")
	}
	if svc.FlyMode() {
		t.Error("Expected initial FlyMode false")
	}

	// Mock connection
	mockPort := &mockSerialPort{}
	svc.mu.Lock()
	svc.port = mockPort
	svc.connected = true
	svc.mu.Unlock()

	// Test GetStatus after connection
	status, _ = svc.GetStatus()
	if !status.Connected {
		t.Error("Expected connected after setting port")
	}

	// Test SetFlymode - 先发送命令，然后模拟设备状态响应（真实设备格式）
	err := svc.SetFlymode(true)
	if err != nil {
		t.Errorf("SetFlymode failed: %v", err)
	}
	// 模拟真实设备的 status_response 格式，flymode 在 mobile 字段下
	// 注意：由于逻辑已取反，如果串口返回 flymode: false，则后端认为 FlyMode 为 true
	simulatedResponse := `{"type":"status_response","version":"1.2.0","mobile":{"number":"","imsi":"123456789012345","uptime":308,"imei":"123456789012345","sim_ready":true,"signal_desc":"强","is_roaming":true,"is_registered":true,"flymode":false,"rsrp":-91,"rsrq":-7,"signal_level":24,"csq":24,"iccid":"12345678901234567890","rssi":-64},"timestamp":1770944358,"mem_kb":57}`
	parsedMsg := &ParsedMessage{
		Type: "status_response",
		JSON: simulatedResponse,
	}
	svc.handleStatusResponse(parsedMsg)

	if !svc.FlyMode() {
		t.Error("Expected FlyMode true after receiving status with flymode=false (inverted logic)")
	}

	// Test RebootMcu (should reset FlyMode)
	err = svc.RebootMcu()
	if err != nil {
		t.Errorf("RebootMcu failed: %v", err)
	}
	if svc.FlyMode() {
		t.Error("Expected FlyMode false after reboot")
	}
}

func TestSerialService_HandleStatusResponse(t *testing.T) {
	logger := zap.NewExample()
	cfg := config.SerialConfig{Port: "/dev/ttyUSB0"}
	svc := NewSerialService(logger, cfg, nil, nil, nil)

	// Simulate receiving a status message with flymode=false (真实设备格式)
	// 由于逻辑已取反，串口返回 false 代表实际开启了飞行模式
	jsonMsg := `{"type":"status_response","version":"1.2.0","mobile":{"number":"","imsi":"123456789012345","uptime":308,"imei":"123456789012345","sim_ready":true,"signal_desc":"强","is_roaming":true,"is_registered":true,"flymode":false,"rsrp":-91,"rsrq":-7,"signal_level":24,"csq":24,"iccid":"12345678901234567890","rssi":-64},"timestamp":1770944358,"mem_kb":57}`
	parsedMsg := &ParsedMessage{
		Type: "status_response",
		JSON: jsonMsg,
	}

	svc.handleStatusResponse(parsedMsg)

	if !svc.FlyMode() {
		t.Error("Expected FlyMode to be true after receiving status with flymode=false (inverted logic)")
	}

	// Simulate receiving a status message with flymode=true (真实设备格式)
	// 由于逻辑已取反，串口返回 true 代表实际关闭了飞行模式
	jsonMsg = `{"type":"status_response","version":"1.2.0","mobile":{"number":"","imsi":"123456789012345","uptime":318,"imei":"123456789012345","sim_ready":true,"signal_desc":"强","is_roaming":true,"is_registered":true,"flymode":true,"rsrp":-91,"rsrq":-7,"signal_level":24,"csq":24,"iccid":"12345678901234567890","rssi":-64},"timestamp":1770944368,"mem_kb":40}`
	parsedMsg = &ParsedMessage{
		Type: "status_response",
		JSON: jsonMsg,
	}

	svc.handleStatusResponse(parsedMsg)

	if svc.FlyMode() {
		t.Error("Expected FlyMode to be false after receiving status with flymode=true (inverted logic)")
	}
}

func TestSerialService_HandleHeartbeatResponse(t *testing.T) {
	logger := zap.NewExample()
	cfg := config.SerialConfig{Port: "/dev/ttyUSB0"}
	svc := NewSerialService(logger, cfg, nil, nil, nil)

	// 测试心跳消息处理（真实设备格式，flymode 在根级别）
	// 由于逻辑已取反，串口返回 flymode: false 代表实际开启了飞行模式
	heartbeatMsg := `{"signal_desc":"强","imei":"123456789012345","rssi":-64,"mem":61,"flymode":false,"type":"heartbeat","signal_level":24,"sim_ready":true,"net_reg":true}`
	// 使用 parseSMSFrame 正确解析心跳消息，确保 Payload 字段被设置
	parsedMsg, err := parseSMSFrame("SMS_START:" + heartbeatMsg + ":SMS_END")
	if err != nil {
		t.Fatalf("Failed to parse heartbeat message: %v", err)
	}

	// 先设置初始状态为 false
	svc.flyMode.Store(false)
	// 处理心跳消息
	svc.handleHeartbeat(parsedMsg)
	// 注意：根据重构逻辑，心跳包不再更新飞行模式，因此这里不再验证 flyMode 状态的变化
}

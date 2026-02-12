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

	// Test SetFlymode
	err := svc.SetFlymode(true)
	if err != nil {
		t.Errorf("SetFlymode failed: %v", err)
	}
	if !svc.FlyMode() {
		t.Error("Expected FlyMode true after setting")
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

	// Simulate receiving a status message with flymode=true
	jsonMsg := `{"type":"status","flymode":true,"mobile":{"imsi":"460001234567890","signal_level":100}}`
	parsedMsg := &ParsedMessage{
		Type: "status",
		JSON: jsonMsg,
	}

	svc.handleStatusResponse(parsedMsg)

	if !svc.FlyMode() {
		t.Error("Expected FlyMode to be true after receiving status with flymode=true")
	}

	// Simulate receiving a status message with flymode=false
	jsonMsg = `{"type":"status","flymode":false,"mobile":{"imsi":"460001234567890","signal_level":100}}`
	parsedMsg = &ParsedMessage{
		Type: "status",
		JSON: jsonMsg,
	}

	svc.handleStatusResponse(parsedMsg)

	if svc.FlyMode() {
		t.Error("Expected FlyMode to be false after receiving status with flymode=false")
	}
}

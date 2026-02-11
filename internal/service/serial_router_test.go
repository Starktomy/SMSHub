package service

import (
	"testing"

	"go.uber.org/zap"
)

func TestSerialServiceMessageRouting(t *testing.T) {
	// Setup SerialService
	logger, _ := zap.NewDevelopment()
	svc := &SerialService{
		logger: logger,
	}

	// Initialize handlers map
	svc.initMessageHandlers()

	if len(svc.handlers) == 0 {
		t.Fatal("Message handlers map is empty")
	}

	// Check if key handlers exist
	requiredHandlers := []string{
		"incoming_sms",
		"heartbeat",
		"incoming_call",
		"sms_send_result",
	}

	for _, h := range requiredHandlers {
		if _, ok := svc.handlers[h]; !ok {
			t.Errorf("Handler for type '%s' is missing", h)
		}
	}

	// Test routing logic (basic)
	// Since we can't easily mock the method receiver (svc) methods without interfaces,
	// we will just ensure routeMessage calls the handler if it exists.
	// But `routeMessage` implementation just does `handler(msg)`.
	// Real integration testing of `routeMessage` requires setting up all dependencies of SerialService
	// (Repo, Notifier, DeviceManager etc) because the handlers use them.
	//
	// However, we can test the negative case: unknown message type

	msg := &ParsedMessage{
		Type: "unknown_type_for_test",
		JSON: "{}",
	}

	// This should log debug message but not panic
	svc.routeMessage(msg)
}

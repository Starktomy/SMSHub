package service

import (
	"context"
	"testing"

	"github.com/Starktomy/smshub/internal/models"
	"go.uber.org/zap"
)

func TestPropertyService(t *testing.T) {
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	svc := NewPropertyService(logger, db)
	ctx := context.Background()

	// 1. Test Set & Get
	err := svc.Set(ctx, "test-key", "Test Property", map[string]string{"foo": "bar"})
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	prop, err := svc.Get(ctx, "test-key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if prop.Name != "Test Property" {
		t.Errorf("Expected name 'Test Property', got '%s'", prop.Name)
	}

	// 2. Test GetValue (Deserialization)
	var val map[string]string
	err = svc.GetValue(ctx, "test-key", &val)
	if err != nil {
		t.Fatalf("GetValue failed: %v", err)
	}
	if val["foo"] != "bar" {
		t.Errorf("Expected value 'bar', got '%s'", val["foo"])
	}

	// 3. Test Cache
	// Modify DB directly to bypass cache update
	db.Model(&models.Property{}).Where("id = ?", "test-key").Update("name", "Modified DB")

	// Should still get old value from cache
	cachedProp, _ := svc.Get(ctx, "test-key")
	if cachedProp.Name != "Test Property" {
		t.Errorf("Cache miss? Expected 'Test Property', got '%s'", cachedProp.Name)
	}

	// 4. Test Set updates/invalidates cache
	err = svc.Set(ctx, "test-key", "Updated Property", map[string]string{"foo": "baz"})
	if err != nil {
		t.Fatalf("Set update failed: %v", err)
	}

	updatedProp, _ := svc.Get(ctx, "test-key")
	if updatedProp.Name != "Updated Property" {
		t.Errorf("Expected 'Updated Property', got '%s'", updatedProp.Name)
	}
}

func TestPropertyServiceInitializeDefaults(t *testing.T) {
	db := setupTestDB(t)
	logger, _ := zap.NewDevelopment()
	svc := NewPropertyService(logger, db)
	ctx := context.Background()

	err := svc.InitializeDefaultConfigs(ctx)
	if err != nil {
		t.Fatalf("InitializeDefaultConfigs failed: %v", err)
	}

	// Verify notification channels config created
	channels, err := svc.GetNotificationChannelConfigs(ctx)
	if err != nil {
		t.Fatalf("GetNotificationChannelConfigs failed: %v", err)
	}
	if channels == nil {
		t.Error("Channels should not be nil")
	}
	if len(channels) != 0 {
		t.Errorf("Expected 0 channels, got %d", len(channels))
	}
}

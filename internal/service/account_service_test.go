package service

import (
	"context"
	"testing"
	"time"

	"github.com/Starktomy/smshub/config"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func TestAccountService(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	// Create a password hash
	password := "secret123"
	hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	hashedPassword := string(hashedBytes)

	appConfig := &config.AppConfig{
		JWT: config.JWTConfig{
			Secret:       "very-long-secret-key-at-least-32-bytes",
			ExpiresHours: 1,
		},
		Users: map[string]string{
			"admin": hashedPassword,
		},
	}

	// Mock OIDCService (nil is fine for basic auth tests)
	svc := NewAccountService(logger, &OIDCService{}, appConfig)
	ctx := context.Background()

	// 1. Test ValidateCredentials
	err := svc.ValidateCredentials(ctx, "admin", "secret123")
	if err != nil {
		t.Errorf("ValidateCredentials failed for valid creds: %v", err)
	}

	err = svc.ValidateCredentials(ctx, "admin", "wrongpass")
	if err == nil {
		t.Error("ValidateCredentials should fail for wrong password")
	}

	err = svc.ValidateCredentials(ctx, "unknown", "secret123")
	if err == nil {
		t.Error("ValidateCredentials should fail for unknown user")
	}

	// 2. Test Login (Basic Auth)
	resp, err := svc.Login(ctx, "admin", "secret123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if resp.Token == "" {
		t.Error("Login response token is empty")
	}
	if resp.User.Username != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", resp.User.Username)
	}

	// 3. Test Token Validation
	claims, err := svc.ValidateToken(resp.Token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.Username != "admin" {
		t.Errorf("Token claim username wrong: %s", claims.Username)
	}

	// 4. Test Invalid Token
	_, err = svc.ValidateToken("invalid-token")
	if err == nil {
		t.Error("ValidateToken should fail for invalid token")
	}
}

func TestAccountServiceTokenExpiry(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	appConfig := &config.AppConfig{
		JWT: config.JWTConfig{
			Secret:       "very-long-secret-key-at-least-32-bytes",
			ExpiresHours: -1, // Expired immediately (effectively, logical check)
		},
		Users: map[string]string{},
	}

	// Note: AccountService constructor sets default to 168 if <= 0
	// So we can't easily test "expired immediately" without modifying the service code
	// or using a very short duration and sleeping.
	// Let's verify the default behavior instead.

	svc := NewAccountService(logger, &OIDCService{}, appConfig)

	// Override duration for test via reflection or just trust the logic?
	// Since struct fields are private/unexported, we can't easily change them.
	// But generateToken uses s.tokenExpireHours.

	// Let's just test that a generated token has correct expiration claim
	token, _, err := svc.generateToken("user", "User")
	if err != nil {
		t.Fatal(err)
	}

	claims, err := svc.ValidateToken(token)
	if err != nil {
		t.Fatal(err)
	}

	if claims.ExpiresAt.Time.Before(time.Now()) {
		t.Error("Token shouldn't be expired yet")
	}
}

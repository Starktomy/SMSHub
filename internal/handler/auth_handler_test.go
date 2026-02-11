package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Starktomy/smshub/config"
	"github.com/Starktomy/smshub/internal/service"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

func setupAuthHandler(t *testing.T) *AuthHandler {
	logger, _ := zap.NewDevelopment()

	// Setup AccountService
	password := "secret"
	hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	appConfig := &config.AppConfig{
		JWT: config.JWTConfig{
			Secret: "secret-key-for-testing-auth-handler-32bytes",
		},
		Users: map[string]string{
			"user": string(hashedBytes),
		},
	}

	accSvc := service.NewAccountService(logger, &service.OIDCService{}, appConfig)
	return NewAuthHandler(logger, accSvc)
}

func TestAuthHandlerLogin(t *testing.T) {
	h := setupAuthHandler(t)
	e := echo.New()

	// 1. Successful Login
	body := `{"username":"user","password":"secret"}`
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Login(c); err != nil {
		t.Fatalf("Login handler failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var resp map[string]interface{}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if _, ok := resp["token"]; !ok {
		t.Error("Response should contain token")
	}

	// 2. Invalid Credentials
	bodyInvalid := `{"username":"user","password":"wrong"}`
	req2 := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(bodyInvalid))
	req2.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec2 := httptest.NewRecorder()
	c2 := e.NewContext(req2, rec2)

	h.Login(c2)
	if rec2.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for invalid creds, got %d", rec2.Code)
	}

	// 3. Missing Fields
	bodyMissing := `{"username":"user"}`
	req3 := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(bodyMissing))
	req3.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec3 := httptest.NewRecorder()
	c3 := e.NewContext(req3, rec3)

	h.Login(c3)
	if rec3.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing fields, got %d", rec3.Code)
	}
}

func TestAuthHandlerGetAuthConfig(t *testing.T) {
	h := setupAuthHandler(t)
	e := echo.New()

	req := httptest.NewRequest(http.MethodGet, "/auth/config", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.GetAuthConfig(c); err != nil {
		t.Fatalf("GetAuthConfig failed: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var config service.AuthConfig
	json.Unmarshal(rec.Body.Bytes(), &config)
	if !config.PasswordEnabled {
		t.Error("PasswordEnabled should be true")
	}
}

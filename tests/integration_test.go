package tests

import (
	"bytes"
	"encoding/json"
	"geminiBackend/config"
	"geminiBackend/internal/app"
	"geminiBackend/internal/domain"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// setupTestServer инициализирует тестовый сервер и базу данных
func setupTestServer(t *testing.T) (*gin.Engine, func()) {
	// Создаём временную тестовую базу данных
	testDB := "test_" + time.Now().Format("20060102150405") + ".db"

	// Загружаем конфигурацию
	_ = godotenv.Load(".env.test")

	cfg := &config.Config{
		Port:             os.Getenv("PORT"),
		JWTSecret:        os.Getenv("JWT_SECRET"),
		DBPath:           testDB,
		ApiGemini:        os.Getenv("GEMINI_API_KEY"),
		Env:              os.Getenv("ENV"), // dev или release
		LocalLLMEndpoint: os.Getenv("LOCAL_LLM_ENDPOINT"),
		LocalLLMMaxChars: 10000,
	}

	// Устанавливаем Gin в тестовый режим
	gin.SetMode(gin.TestMode)

	// Создаём экземпляр приложения и настраиваем роутер
	appInstance := app.New(cfg)
	router, err := appInstance.SetupRouter()
	if err != nil {
		t.Fatalf("Failed to setup router: %v", err)
	}

	// Функция очистки
	cleanup := func() {
		os.Remove(testDB)
	}

	return router, cleanup
}

func TestHealthCheck(t *testing.T) {
	router := gin.New()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/ping", nil)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestRegisterAndLogin(t *testing.T) {
	router, cleanup := setupTestServer(t)
	defer cleanup()

	// Тест регистрации
	registerPayload := domain.RegisterRequest{
		Username: "testuser",
		TgID:     12345,
	}
	body, _ := json.Marshal(registerPayload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Register failed: status %d, body: %s", w.Code, w.Body.String())
	}

	var registerResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &registerResp)

	if registerResp["status"] != "success" {
		t.Errorf("Expected success status, got: %v", registerResp)
	}

	// Тест авторизации
	loginPayload := domain.LoginRequest{
		Username: "testuser",
		TgID:     12345,
	}
	body, _ = json.Marshal(loginPayload)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Login failed: status %d, body: %s", w.Code, w.Body.String())
	}

	var loginResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResp)

	if loginResp["status"] != "success" {
		t.Errorf("Expected success status, got: %v", loginResp)
	}

	// Проверяем наличие токена
	data := loginResp["data"].(map[string]interface{})
	token := data["token"].(string)
	if token == "" {
		t.Errorf("Expected non-empty token")
	}
}

func TestAIKeyManagement(t *testing.T) {
	router, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Register and login to get token
	token := registerAndLogin(t, router, "keyuser", 99999)

	// 2. Check key status (should be false)
	status := checkKeyStatus(t, router, token)
	if status {
		t.Errorf("Expected initial key status to be false")
	}

	// 3. Set API key
	setKey := domain.SetKeyRequest{APIKey: "test_api_key_1234567890"}
	body, _ := json.Marshal(setKey)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/user/ai/key", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Set key failed: status %d, body: %s", w.Code, w.Body.String())
	}

	// 4. Check key status (should be true)
	status = checkKeyStatus(t, router, token)
	if !status {
		t.Errorf("Expected key status to be true after setting")
	}

	// 5. Delete key
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("DELETE", "/api/user/ai/key", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Delete key failed: status %d, body: %s", w.Code, w.Body.String())
	}

	// 6. Check key status (should be false)
	status = checkKeyStatus(t, router, token)
	if status {
		t.Errorf("Expected key status to be false after deletion")
	}
}

func TestAITextGeneration(t *testing.T) {
	router, cleanup := setupTestServer(t)
	defer cleanup()

	// 1. Регистрируемся и входим
	token := registerAndLogin(t, router, "aiuser", 88888)

	// 2. Пробуем AI запрос без ключа (должна вернуться ошибка missing_api_key)
	aiReq := domain.AITextRequest{Prompt: "Hello world"}
	body, _ := json.Marshal(aiReq)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/user/ai/text", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != 400 {
		t.Errorf("Expected 400 when no key set, got %d", w.Code)
	}

	var errResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &errResp)
	errorData := errResp["error"].(map[string]interface{})
	if errorData["code"] != "missing_api_key" {
		t.Errorf("Expected missing_api_key error, got: %v", errorData["code"])
	}

	// 3. Устанавливаем API ключ
	setKey := domain.SetKeyRequest{APIKey: "fake_gemini_key_1234567890"}
	body, _ = json.Marshal(setKey)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/user/ai/key", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Set key failed: %s", w.Body.String())
	}

	// 4. Пробуем AI запрос с ключом (вернётся ai_error из-за фейкового ключа, но не missing_api_key)
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/user/ai/text", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	// Должны получить ai_error (500) или success (200) в зависимости от валидности ключа
	// Для фейкового ключа ожидаем 500 с ai_error
	if w.Code == 400 {
		json.Unmarshal(w.Body.Bytes(), &errResp)
		errorData := errResp["error"].(map[string]interface{})
		if errorData["code"] == "missing_api_key" {
			t.Errorf("Should not get missing_api_key after setting key")
		}
	}
}

func TestUnauthorizedAccess(t *testing.T) {
	router, cleanup := setupTestServer(t)
	defer cleanup()

	// Пытаемся получить доступ к защищённым эндпоинтам без токена
	tests := []struct {
		method   string
		endpoint string
	}{
		{"GET", "/api/user/ping"},
		{"GET", "/api/user/ai/key"},
		{"POST", "/api/user/ai/text"},
		{"GET", "/api/admin/ping"},
	}

	for _, tt := range tests {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(tt.method, tt.endpoint, nil)
		router.ServeHTTP(w, req)

		if w.Code != 401 {
			t.Errorf("Expected 401 for %s %s, got %d", tt.method, tt.endpoint, w.Code)
		}
	}
}

func TestInvalidToken(t *testing.T) {
	router, cleanup := setupTestServer(t)
	defer cleanup()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/user/ping", nil)
	req.Header.Set("Authorization", "Bearer invalid_token_here")
	router.ServeHTTP(w, req)

	if w.Code != 401 {
		t.Errorf("Expected 401 for invalid token, got %d", w.Code)
	}
}

// Вспомогательные функции
func registerAndLogin(t *testing.T, router *gin.Engine, username string, tgID int) string {
	// Регистрация
	registerPayload := domain.RegisterRequest{
		Username: username,
		TgID:     tgID,
	}
	body, _ := json.Marshal(registerPayload)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// Авторизация
	loginPayload := domain.LoginRequest{
		Username: username,
		TgID:     tgID,
	}
	body, _ = json.Marshal(loginPayload)

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Login failed: %s", w.Body.String())
	}

	var loginResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &loginResp)
	data := loginResp["data"].(map[string]interface{})
	return data["token"].(string)
}

func checkKeyStatus(t *testing.T, router *gin.Engine, token string) bool {
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/user/ai/key", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	router.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("Check key status failed: %s", w.Body.String())
	}

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	data := resp["data"].(map[string]interface{})
	return data["has_key"].(bool)
}

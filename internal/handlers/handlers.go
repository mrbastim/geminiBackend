package handlers

import (
	"encoding/json"
	"geminiBackend/config"
	"geminiBackend/pkg/utils"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Обработка получения опций от сервера
func GetOptions(w http.ResponseWriter, r *http.Request, config *config.Config) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// Логика обработки GET-запроса и отправки опций клиенту
	payload := map[string]interface{}{
		"Port":   config.Port,
		"apiKey": config.ApiGemini,
	}
	utils.RespondWithJSON(w, http.StatusOK, payload)
}

// Получение ответа от Gemini API и отправка его клиенту
func PostResponse(w http.ResponseWriter, r *http.Request) {

}

// LoginRequest структура для запроса авторизации
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse структура для ответа с токеном
type LoginResponse struct {
	Token string `json:"token"`
}

// Claims структура для JWT
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// Login обрабатывает авторизацию и выдает JWT токен
func Login(w http.ResponseWriter, r *http.Request, config *config.Config) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Простая проверка учетных данных (в продакшене используйте БД и хеширование)
	var role string
	if req.Username == "admin" && req.Password == "admin123" {
		role = "admin"
	} else if req.Username == "user" && req.Password == "user123" {
		role = "user"
	} else {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Создаем JWT токен
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		Username: req.Username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(config.JWTSecret))
	if err != nil {
		http.Error(w, "Error generating token", http.StatusInternalServerError)
		return
	}

	response := LoginResponse{Token: tokenString}
	utils.RespondWithJSON(w, http.StatusOK, response)
	log.Printf("User %s logged in with role %s", req.Username, role)
}

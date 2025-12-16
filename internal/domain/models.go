package domain

import (
	"database/sql"

	"github.com/golang-jwt/jwt/v5"
)

type LoginRequest struct {
	Username string `json:"username"`
	TgID     int    `json:"tg_id"`
}

type RegisterRequest struct {
	Username string `json:"username"`
	TgID     int    `json:"tg_id"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterResponse struct {
	Message string `json:"message"`
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	TgID     int    `json:"tg_id"`
	jwt.RegisteredClaims
}

type User struct {
	Username string
	Role     string
}

type UserDB struct {
	ID           int64
	TgID         int
	Username     string
	GeminiAPIKey sql.NullString
	IsAdmin      int
	IsActive     int
	LastLogin    sql.NullTime
	CreatedAt    sql.NullTime
	UpdatedAt    sql.NullTime
}

type SetKeyRequest struct {
	APIKey string `json:"api_key"`
}

type KeyStatusResponse struct {
	HasKey bool `json:"has_key"`
}

// ModelInfo информация о модели Gemini
type ModelInfo struct {
	Name             string   `json:"name"`               // Имя модели (например, "gemini-2.5-flash")
	DisplayName      string   `json:"display_name"`       // Отображаемое имя
	Description      string   `json:"description"`        // Описание модели
	SupportedActions []string `json:"supported_actions"`  // Поддерживаемые действия (generateContent, etc)
	InputTokenLimit  int32    `json:"input_token_limit"`  // Лимит входных токенов
	OutputTokenLimit int32    `json:"output_token_limit"` // Лимит выходных токенов
	Category         string   `json:"category"`           // Категория: text, multimodal, embedding, etc
	IsAvailable      bool     `json:"is_available"`       // Доступна ли модель сейчас
}

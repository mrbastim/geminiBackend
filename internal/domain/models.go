package domain

import (
	"database/sql"

	"github.com/golang-jwt/jwt/v5"
)

type LoginRequest struct {
	Username string `json:"username"`
	TgID     int    `json:"tg_id"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
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

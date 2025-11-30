package service

import (
	"database/sql"
	"geminiBackend/internal/domain"
	"geminiBackend/internal/provider/db"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	jwtSecret string
}

func NewAuthService(secret string) *AuthService {
	return &AuthService{jwtSecret: secret}
}

func (s *AuthService) Login(req domain.LoginRequest) (domain.LoginResponse, error) {
	var role string

	sqlDB, err := sql.Open("sqlite3", "data/gemini.db")
	if err != nil {
		return domain.LoginResponse{}, err
	}
	userDB := db.NewUsersProvider(sqlDB)
	defer userDB.Close()

	user, err := userDB.GetUserByTelegramID(req.TgID)
	if err != nil {
		return domain.LoginResponse{}, domain.ErrInvalidCredentials
	}

	isAdmin := user.IsAdmin
	switch {
	case isAdmin:
		role = "admin"
	case !isAdmin:
		role = "user"
	default:
		return domain.LoginResponse{}, domain.ErrInvalidCredentials
	}
	exp := time.Now().Add(1 * time.Hour)
	claims := &domain.Claims{Username: req.Username, Role: role, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(exp), IssuedAt: jwt.NewNumericDate(time.Now())}}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return domain.LoginResponse{}, err
	}
	return domain.LoginResponse{Token: tokenString}, nil
}

func (s *AuthService) Parse(tokenString string) (*domain.Claims, error) {
	claims := &domain.Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, domain.ErrUnauthorized
	}
	return claims, nil
}

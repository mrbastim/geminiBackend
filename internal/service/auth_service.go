package service

import (
	"database/sql"
	"geminiBackend/internal/domain"
	"geminiBackend/internal/provider/db"
	"geminiBackend/pkg/logger"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	jwtSecret string
	db        *sql.DB
}

func NewAuthService(secret string, database *sql.DB) *AuthService {
	return &AuthService{jwtSecret: secret, db: database}
}

func (s *AuthService) Login(req domain.LoginRequest) (domain.LoginResponse, error) {
	if req.TgID <= 0 {
		return domain.LoginResponse{}, domain.ErrInvalidCredentials
	}

	userDB := db.NewUsersProvider(s.db)

	user, err := userDB.GetUserByTelegramID(req.TgID)
	if err != nil {
		logger.L.Error("get user error", "tg_id", req.TgID, "err", err)
		return domain.LoginResponse{}, domain.ErrInvalidCredentials
	}

	if user.IsActive == 0 {
		return domain.LoginResponse{}, domain.ErrInvalidCredentials
	}

	var role string
	if user.IsAdmin == 1 {
		role = "admin"
	} else {
		role = "user"
	}
	exp := time.Now().Add(1 * time.Hour)
	claims := &domain.Claims{Username: req.Username, Role: role, TgID: user.TgID, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(exp), IssuedAt: jwt.NewNumericDate(time.Now())}}
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

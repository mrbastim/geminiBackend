package service

import (
	"geminiBackend/internal/domain"
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
	switch {
	case req.Username == "admin" && req.Password == "admin123":
		role = "admin"
	case req.Username == "user" && req.Password == "user123":
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

package domain

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidInput       = errors.New("invalid input")
)

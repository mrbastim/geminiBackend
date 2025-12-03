package main

// @title Gemini Backend API
// @version 1.0
// @description REST API для взаимодействия с Gemini AI и аутентификации.
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

import (
	"geminiBackend/config"
	_ "geminiBackend/docs"
	"geminiBackend/internal/app"
	"geminiBackend/pkg/logger"
)

func main() {
	cfg, err := config.LoadConfig("config/config.yaml")
	logger.L.Info("loading config")
	if err != nil {
		logger.L.Error("config load error", "err", err)
		return
	}
	if cfg.JWTSecret == "CHANGE_ME" {
		logger.L.Warn("default JWT secret in use")
	}

	// DB schema will be initialized in app.Run

	application := app.New(cfg)
	if err := application.Run(); err != nil {
		logger.L.Error("server error", "err", err)
	}
}

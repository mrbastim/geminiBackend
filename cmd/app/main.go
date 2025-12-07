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
	cfg := config.LoadConfig()
	logger.L.Info("loading config", "env", cfg.Env)
	logger.Init(cfg.LogLevel, cfg.LogFile)

	if cfg.JWTSecret == "change-me-in-production" {
		logger.L.Warn("warning: default JWT secret in use")
	}
	if cfg.ApiGemini == "" {
		logger.L.Warn("warning: GEMINI_API_KEY not set")
	}

	logger.L.Info("starting application", "env", cfg.Env, "port", cfg.Port)

	application := app.New(cfg)
	if err := application.Run(); err != nil {
		logger.L.Error("server error", "err", err)
	}
}

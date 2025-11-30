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
	"log"
)

func main() {
	cfg, err := config.LoadConfig("config/config.yaml")
	log.Println("loading config")
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}
	if cfg.JWTSecret == "CHANGE_ME" {
		log.Println("Warning: default JWT secret in use")
	}
	application := app.New(cfg)
	if err := application.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

package main

import (
	"geminiBackend/config"
	"geminiBackend/internal/app"
	"log"
)

func main() {
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("config load error: %v", err)
	}
	if cfg.JWTSecret == "CHANGE_ME" {
		log.Println("warning: default JWT secret in use")
	}
	application := app.New(cfg)
	if err := application.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

package app

import (
	"geminiBackend/config"
	delivery "geminiBackend/internal/delivery/http"
	"geminiBackend/internal/delivery/http/middleware"
	"geminiBackend/internal/provider/db"
	"geminiBackend/internal/provider/gemini"
	"geminiBackend/internal/service"
	"log"
	"net/http"
	"os"
)

type App struct {
	cfg *config.Config
}

func New(cfg *config.Config) *App { return &App{cfg: cfg} }

func (a *App) Run() error {
	// Init SQLite DB and ensure schema
	dbPath := a.cfg.DBPath
	if dbPath == "" {
		dbPath = "data.db"
	}
	sqlDB, err := db.InitDBLite(dbPath)
	if err != nil {
		return err
	}
	defer sqlDB.Close()
	// You may want to close DB on shutdown; omitted for brevity

	// Providers and services
	authService := service.NewAuthService(a.cfg.JWTSecret)
	apiKey := os.Getenv("API_GEMINI_KEY")
	gemClient := gemini.NewClient(apiKey)
	aiService := service.NewAIService(gemClient)
	handler := delivery.NewHandler(authService, aiService)
	router := delivery.NewRouter(handler, middleware.JWT(authService), middleware.RequireAdmin)
	log.Printf("starting server on :%s", a.cfg.Port)
	return http.ListenAndServe(":"+a.cfg.Port, router)
}

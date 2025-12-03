package app

import (
	"context"
	"geminiBackend/config"
	delivery "geminiBackend/internal/delivery/http"
	"geminiBackend/internal/delivery/http/middleware"
	"geminiBackend/internal/provider/db"
	"geminiBackend/internal/service"
	"geminiBackend/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
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
	authService := service.NewAuthService(a.cfg.JWTSecret, sqlDB)
	// AI service now uses per-user API keys
	aiService := service.NewAIService()
	handler := delivery.NewHandler(authService, aiService, sqlDB)
	// Rate limiters
	rl := middleware.NewIPRateLimiter(10, time.Minute) // 10 requests per minute per IP
	router := delivery.NewRouter(handler, middleware.JWT(authService), middleware.RequireAdmin, rl)

	srv := &http.Server{Addr: ":" + a.cfg.Port, Handler: router}
	// Start server in goroutine
	go func() {
		logger.L.Info("starting server", "port", a.cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.L.Error("listen error", "err", err)
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.L.Info("shutting down server")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.L.Error("server shutdown error", "err", err)
	}
	return nil
}

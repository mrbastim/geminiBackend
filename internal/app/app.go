package app

import (
	"database/sql"
	"geminiBackend/config"
	delivery "geminiBackend/internal/delivery/http"
	"geminiBackend/internal/delivery/http/middleware"
	"geminiBackend/internal/provider/db"
	"geminiBackend/internal/service"
	"geminiBackend/pkg/logger"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

type App struct {
	cfg    *config.Config
	router *gin.Engine
	sqlDB  *sql.DB
}

func New(cfg *config.Config) *App { return &App{cfg: cfg} }

func (a *App) SetupRouter() (*gin.Engine, error) {
	dbPath := a.cfg.DBPath
	if dbPath == "" {
		dbPath = "data.db"
	}
	sqlDB, err := db.InitDBLite(dbPath)
	if err != nil {
		return nil, err
	}
	a.sqlDB = sqlDB

	// Провайдеры и сервисы
	authService := service.NewAuthService(a.cfg.JWTSecret, sqlDB)
	aiService := service.NewAIService()
	handler := delivery.NewHandler(authService, aiService, sqlDB)

	// Rate limiters
	rl := middleware.NewIPRateLimiter(10, time.Minute)

	// Gin роутер
	ginRouter := delivery.NewRouter(handler, middleware.JWTAuth(authService), middleware.AdminOnly(), rl)
	a.router = ginRouter
	return ginRouter, nil
}

func (a *App) Run() error {
	ginRouter, err := a.SetupRouter()
	if err != nil {
		return err
	}
	defer a.sqlDB.Close()

	// Установка режима Gin (debug/release) через env
	if a.cfg.GinMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Установка доверенных proxies
	if len(a.cfg.TrustedProxies) > 0 {
		ginRouter.SetTrustedProxies(a.cfg.TrustedProxies)
	} else {
		// По умолчанию доверяем localhost
		ginRouter.SetTrustedProxies([]string{"127.0.0.1", "localhost"})
	}

	// Запускаем сервер в горутине
	go func() {
		logger.L.Info("starting server", "port", a.cfg.Port, "mode", a.cfg.Env)
		if err := ginRouter.Run(":" + a.cfg.Port); err != nil {
			logger.L.Error("listen error", "err", err)
		}
	}()

	// Корректное завершение работы
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.L.Info("shutting down server")

	return nil
}

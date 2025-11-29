package app

import (
	"geminiBackend/config"
	delivery "geminiBackend/internal/delivery/http"
	"geminiBackend/internal/delivery/http/middleware"
	"geminiBackend/internal/service"
	"log"
	"net/http"
)

type App struct {
	cfg *config.Config
}

func New(cfg *config.Config) *App { return &App{cfg: cfg} }

func (a *App) Run() error {
	authService := service.NewAuthService(a.cfg.JWTSecret)
	handler := delivery.NewHandler(authService)
	router := delivery.NewRouter(handler, middleware.JWT(authService), middleware.RequireAdmin)
	log.Printf("starting server on :%s", a.cfg.Port)
	return http.ListenAndServe(":"+a.cfg.Port, router)
}

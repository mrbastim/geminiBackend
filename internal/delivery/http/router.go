package http

import (
	"geminiBackend/internal/delivery/http/middleware"
	"geminiBackend/pkg/logger"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func NewRouter(h *Handler, jwtMiddleware gin.HandlerFunc, adminOnly gin.HandlerFunc, rl middleware.RateLimiter) *gin.Engine {
	r := gin.Default()
	logger.L.Info("Initializing Gin router")

	api := r.Group("/api")

	// Публичные маршруты с rate limiting
	public := api.Group("")
	public.Use(rl.GinMiddleware())
	public.POST("/login", h.Login)
	public.POST("/register", h.Register)

	// Маршруты администратора
	admin := api.Group("/admin")
	admin.Use(jwtMiddleware, adminOnly)
	admin.GET("/ping", h.AdminPing)
	admin.GET("/options", h.Options)

	// Пользовательские маршруты
	user := api.Group("/user")
	user.Use(jwtMiddleware)
	user.GET("/ping", h.UserPing)
	user.POST("/ai/text", rl.GinMiddleware(), h.AIText)
	user.POST("/ai/key", rl.GinMiddleware(), h.AISetKey)
	user.DELETE("/ai/key", rl.GinMiddleware(), h.AIClearKey)
	user.GET("/ai/key", rl.GinMiddleware(), h.AIKeyStatus)

	// Swagger документация
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler()))

	return r
}

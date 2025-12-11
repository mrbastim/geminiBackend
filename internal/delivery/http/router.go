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

	var rlMiddleware gin.HandlerFunc
	if rl != nil {
		rlMiddleware = rl.GinMiddleware()
	} else {
		rlMiddleware = func(c *gin.Context) {
			c.Next()
		}
	}

	api := r.Group("/api")

	// Публичные маршруты с rate limiting
	public := api.Group("")
	public.Use(rlMiddleware)
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
	user.GET("/ai/models", rlMiddleware, h.AIModels)
	user.POST("/ai/text", rlMiddleware, h.AIText)
	user.POST("/ai/key", rlMiddleware, h.AISetKey)
	user.DELETE("/ai/key", rlMiddleware, h.AIClearKey)
	user.GET("/ai/key", rlMiddleware, h.AIKeyStatus)

	// Swagger документация
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.NewHandler()))

	return r
}

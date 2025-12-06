package middleware

import (
	"geminiBackend/internal/domain"
	"geminiBackend/internal/service"
	"geminiBackend/pkg/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const ClaimsContextKey = "claims"

func JWTAuth(auth *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		parts := strings.Split(header, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Error(c.Writer, http.StatusUnauthorized, "auth_error", "invalid auth header format")
			c.Abort()
			return
		}

		claims, err := auth.Parse(parts[1])
		if err != nil {
			utils.Error(c.Writer, http.StatusUnauthorized, "invalid_token", "invalid or expired token")
			c.Abort()
			return
		}

		c.Set(ClaimsContextKey, claims)
		c.Next()
	}
}

func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get(ClaimsContextKey)
		if !exists {
			utils.Error(c.Writer, http.StatusForbidden, "no_claims", "claims not found in context")
			c.Abort()
			return
		}

		userClaims, ok := claims.(*domain.Claims)
		if !ok || userClaims.Role != "admin" {
			utils.Error(c.Writer, http.StatusForbidden, "forbidden", "admin access required")
			c.Abort()
			return
		}

		c.Next()
	}
}

// ClaimsFromContext возвращает клеймы из Gin контекста
func ClaimsFromContext(c *gin.Context) (*domain.Claims, bool) {
	claims, exists := c.Get(ClaimsContextKey)
	if !exists {
		return nil, false
	}
	userClaims, ok := claims.(*domain.Claims)
	return userClaims, ok
}

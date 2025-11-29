package middleware

import (
	"context"
	"geminiBackend/internal/domain"
	"geminiBackend/internal/service"
	"net/http"
	"strings"
)

type claimsKey struct{}

func JWT(auth *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			parts := strings.Split(header, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				http.Error(w, "invalid auth header", http.StatusUnauthorized)
				return
			}
			claims, err := auth.Parse(parts[1])
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), claimsKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, ok := r.Context().Value(claimsKey{}).(*domain.Claims)
		if !ok || c.Role != "admin" {
			http.Error(w, "admin required", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

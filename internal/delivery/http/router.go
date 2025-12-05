package http

import (
	stdhttp "net/http"

	"geminiBackend/internal/delivery/http/middleware"

	"github.com/gorilla/mux"
	httpSwagger "github.com/swaggo/http-swagger"
)

func NewRouter(h *Handler, jwtMiddleware func(stdhttp.Handler) stdhttp.Handler, adminOnly func(stdhttp.Handler) stdhttp.Handler, rl middleware.RateLimiter) *mux.Router {
	r := mux.NewRouter().PathPrefix("/api").Subrouter()
	r.Handle("/login", rl.Limit(stdhttp.HandlerFunc(h.Login))).Methods("POST")
	r.Handle("/register", rl.Limit(stdhttp.HandlerFunc(h.Register))).Methods("POST")
	// protected example
	secure := r.PathPrefix("/admin").Subrouter()
	secure.Use(jwtMiddleware, adminOnly)
	secure.HandleFunc("/ping", func(w stdhttp.ResponseWriter, r *stdhttp.Request) { w.Write([]byte("admin pong")) }).Methods("GET")
	secure.HandleFunc("/options", h.Options).Methods("GET")

	user := r.PathPrefix("/user").Subrouter()
	user.Use(jwtMiddleware)
	user.HandleFunc("/ping", func(w stdhttp.ResponseWriter, r *stdhttp.Request) { w.Write([]byte("user pong")) }).Methods("GET")
	user.Handle("/ai/text", rl.Limit(stdhttp.HandlerFunc(h.AIText))).Methods("POST")
	user.Handle("/ai/key", rl.Limit(stdhttp.HandlerFunc(h.AISetKey))).Methods("POST")
	user.Handle("/ai/key", rl.Limit(stdhttp.HandlerFunc(h.AIClearKey))).Methods("DELETE")
	user.Handle("/ai/key", rl.Limit(stdhttp.HandlerFunc(h.AIKeyStatus))).Methods("GET")

	// Swagger UI (после генерации документации командой `make swagger`)
	r.PathPrefix("/swagger/").Handler(httpSwagger.WrapHandler)

	return r
}

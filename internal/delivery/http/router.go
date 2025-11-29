package http

import (
	stdhttp "net/http"

	"github.com/gorilla/mux"
)

func NewRouter(h *Handler, jwtMiddleware func(stdhttp.Handler) stdhttp.Handler, adminOnly func(stdhttp.Handler) stdhttp.Handler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/login", h.Login).Methods("POST")
	// protected example
	secure := r.PathPrefix("/admin").Subrouter()
	secure.Use(jwtMiddleware, adminOnly)
	secure.HandleFunc("/ping", func(w stdhttp.ResponseWriter, r *stdhttp.Request) { w.Write([]byte("admin pong")) }).Methods("GET")
	secure.HandleFunc("/options", h.Options).Methods("GET")

	user := r.PathPrefix("/user").Subrouter()
	user.Use(jwtMiddleware)
	user.HandleFunc("/ping", func(w stdhttp.ResponseWriter, r *stdhttp.Request) { w.Write([]byte("user pong")) }).Methods("GET")
	user.HandleFunc("/ai/text", h.AIText).Methods("POST")

	return r
}

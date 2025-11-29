package http

import (
	"geminiBackend/pkg/utils"
	stdhttp "net/http"

	"github.com/gorilla/mux"
)

func NewRouter(h *Handler, jwtMiddleware func(stdhttp.Handler) stdhttp.Handler, adminOnly func(stdhttp.Handler) stdhttp.Handler) *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/login", h.Login).Methods("POST")

	secure := r.PathPrefix("/admin").Subrouter()
	secure.Use(jwtMiddleware, adminOnly)
	secure.HandleFunc("/ping", func(w stdhttp.ResponseWriter, r *stdhttp.Request) {
		utils.RespondSuccess(w, 200, map[string]string{"message": "admin pong"})
	}).Methods("GET")
	secure.HandleFunc("/options", h.Options).Methods("GET")

	return r
}

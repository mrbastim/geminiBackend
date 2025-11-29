package http

import (
	"encoding/json"
	"geminiBackend/config"
	"geminiBackend/internal/domain"
	"geminiBackend/internal/service"
	"geminiBackend/pkg/utils"
	"log"
	"net/http"
)

type Handler struct {
	auth *service.AuthService
}

func NewHandler(auth *service.AuthService) *Handler { return &Handler{auth: auth} }

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.RespondError(w, http.StatusBadRequest, "bad_request", "invalid body")
		return
	}
	resp, err := h.auth.Login(req)
	if err != nil {
		utils.RespondError(w, http.StatusUnauthorized, "invalid_credentials", err.Error())
		return
	}
	utils.RespondSuccess(w, http.StatusOK, resp)
	log.Printf("user %s logged in", req.Username)
}

func (h *Handler) Options(w http.ResponseWriter, r *http.Request) {
	config, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		utils.RespondError(w, http.StatusInternalServerError, "config_error", "failed to load config")
		return
	}
	utils.RespondSuccess(w, http.StatusOK, map[string]string{"port": config.Port})
}

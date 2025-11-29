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
	ai   *service.AIService
}

func NewHandler(auth *service.AuthService, ai *service.AIService) *Handler {
	return &Handler{auth: auth, ai: ai}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req domain.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "bad_request", "invalid body")
		return
	}
	resp, err := h.auth.Login(req)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}
	utils.Success(w, resp)
	log.Printf("user %s logged in", req.Username)
}

func (h *Handler) Options(w http.ResponseWriter, r *http.Request) {
	config, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "config_error", "failed to load config")
		return
	}
	utils.Success(w, map[string]string{"port": config.Port})
}

type aiTextRequest struct {
	Prompt string `json:"prompt"`
}

func (h *Handler) AIText(w http.ResponseWriter, r *http.Request) {
	var req aiTextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "bad_request", "invalid body")
		return
	}
	if req.Prompt == "" {
		utils.Error(w, http.StatusBadRequest, "validation_error", "prompt required")
		return
	}
	text, err := h.ai.AskText(req.Prompt)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "ai_error", err.Error())
		return
	}
	utils.Success(w, map[string]string{"text": text})
}

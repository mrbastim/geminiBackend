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

// @Summary Логин
// @Description Аутентификация пользователя и получение JWT токена
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body domain.LoginRequest true "Данные для входа"
// @Success 200 {object} domain.LoginSuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Router /login [post]
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

// @Summary Опции сервера
// @Description Возвращает основные параметры конфигурации (пример)
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.OptionsSuccessResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /admin/options [get]
func (h *Handler) Options(w http.ResponseWriter, r *http.Request) {
	config, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "config_error", "failed to load config")
		return
	}
	utils.Success(w, map[string]string{"port": config.Port})
}

// @Summary Генерация текста
// @Description Генерирует текст по переданному prompt через Gemini
// @Tags ai
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body domain.AITextRequest true "Запрос на генерацию"
// @Success 200 {object} domain.AITextSuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /user/ai/text [post]
func (h *Handler) AIText(w http.ResponseWriter, r *http.Request) {
	var req domain.AITextRequest
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

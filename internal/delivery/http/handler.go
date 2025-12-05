package http

import (
	"database/sql"
	"encoding/json"
	"geminiBackend/config"
	"geminiBackend/internal/delivery/http/middleware"
	"geminiBackend/internal/domain"
	"geminiBackend/internal/provider/db"
	"geminiBackend/internal/service"
	"geminiBackend/pkg/logger"
	"geminiBackend/pkg/utils"
	"net/http"
)

type Handler struct {
	auth *service.AuthService
	ai   *service.AIService
	db   *sql.DB
}

func NewHandler(auth *service.AuthService, ai *service.AIService, database *sql.DB) *Handler {
	return &Handler{auth: auth, ai: ai, db: database}
}

// @Summary Регистрация
// @Description Регистрация нового пользователя
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body domain.RegisterRequest true "Данные для регистрации"
// @Success 200 {object} domain.RegisterResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Router /register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req domain.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "bad_request", "invalid body")
		return
	}
	resp, err := h.auth.Register(req)
	if err != nil {
		if err == domain.ErrUserExists {
			utils.Error(w, http.StatusBadRequest, "user_exists", err.Error())
			return
		}
		utils.Error(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	utils.Success(w, resp)
	logger.L.Debug("new user registered", "username", req.Username)
}

// @Summary Логин
// @Description Аутентификация пользователя и получение JWT токена
// @Tags auth
// @Accept json
// @Produce json
// @Param payload body domain.LoginRequest true "Данные для входа"
// @Success 200 {object} domain.LoginResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 429 {object} domain.ErrorResponse
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
	logger.L.Debug("user logged in", "username", req.Username)
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
// @Failure 429 {object} domain.ErrorResponse
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
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized", "no claims")
		return
	}
	users := db.NewUsersProvider(h.db)
	user, err := users.GetUserByTelegramID(claims.TgID)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthorized", "user not found")
		return
	}
	if !user.GeminiAPIKey.Valid || user.GeminiAPIKey.String == "" {
		utils.Error(w, http.StatusBadRequest, "missing_api_key", "set your Gemini API key first")
		return
	}
	text, err := h.ai.AskText(user.GeminiAPIKey.String, req.Prompt)
	if err != nil {
		utils.Error(w, http.StatusInternalServerError, "ai_error", err.Error())
		return
	}
	utils.Success(w, map[string]string{"text": text})
}

// @Summary Установить ключ Gemini
// @Description Сохраняет пользовательский Gemini API ключ (перезаписывает существующий)
// @Tags ai
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payload body domain.SetKeyRequest true "Ключ Gemini"
// @Success 200 {object} domain.OptionsSuccessResponse
// @Failure 400 {object} domain.ErrorResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 429 {object} domain.ErrorResponse
// @Router /user/ai/key [post]
func (h *Handler) AISetKey(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized", "no claims")
		return
	}
	var req domain.SetKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.Error(w, http.StatusBadRequest, "bad_request", "invalid body")
		return
	}
	if len(req.APIKey) < 10 { // простая валидация длины
		utils.Error(w, http.StatusBadRequest, "validation_error", "api_key too short")
		return
	}
	users := db.NewUsersProvider(h.db)
	if err := users.SetGeminiAPIKey(claims.TgID, req.APIKey); err != nil {
		utils.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	logger.L.Debug("user set gemini key", "tg_id", claims.TgID)
	utils.Success(w, map[string]string{"status": "ok"})
}

// @Summary Удалить ключ Gemini
// @Tags ai
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.OptionsSuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Router /user/ai/key [delete]
func (h *Handler) AIClearKey(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized", "no claims")
		return
	}
	users := db.NewUsersProvider(h.db)
	if err := users.ClearGeminiAPIKey(claims.TgID); err != nil {
		utils.Error(w, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	logger.L.Debug("user cleared gemini key", "tg_id", claims.TgID)
	utils.Success(w, map[string]string{"status": "ok"})
}

// @Summary Статус ключа Gemini
// @Tags ai
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.KeyStatusResponse
// @Failure 401 {object} domain.ErrorResponse
// @Router /user/ai/key [get]
func (h *Handler) AIKeyStatus(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.ClaimsFromContext(r.Context())
	if !ok {
		utils.Error(w, http.StatusUnauthorized, "unauthorized", "no claims")
		return
	}
	users := db.NewUsersProvider(h.db)
	user, err := users.GetUserByTelegramID(claims.TgID)
	if err != nil {
		utils.Error(w, http.StatusUnauthorized, "unauthorized", "user not found")
		return
	}
	utils.Success(w, domain.KeyStatusResponse{HasKey: user.GeminiAPIKey.Valid && user.GeminiAPIKey.String != ""})
}

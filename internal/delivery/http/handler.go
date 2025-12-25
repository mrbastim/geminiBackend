package http

import (
	"database/sql"
	"geminiBackend/internal/delivery/http/middleware"
	"geminiBackend/internal/domain"
	"geminiBackend/internal/provider/db"
	"geminiBackend/internal/provider/gemini"
	"geminiBackend/internal/service"
	"geminiBackend/pkg/logger"
	"geminiBackend/pkg/utils"
	"net/http"

	"github.com/gin-gonic/gin"
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
func (h *Handler) Register(c *gin.Context) {
	var req domain.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, http.StatusBadRequest, "bad_request", "invalid body")
		return
	}
	resp, err := h.auth.Register(req)
	if err != nil {
		if err == domain.ErrUserExists {
			utils.Error(c.Writer, http.StatusBadRequest, "user_exists", err.Error())
			return
		}
		utils.Error(c.Writer, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	utils.Success(c.Writer, resp)
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
func (h *Handler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, http.StatusBadRequest, "bad_request", "invalid body")
		return
	}
	resp, err := h.auth.Login(req)
	if err != nil {
		utils.Error(c.Writer, http.StatusUnauthorized, "unauthorized", err.Error())
		return
	}
	utils.Success(c.Writer, resp)
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
func (h *Handler) Options(c *gin.Context) {
	utils.Success(c.Writer, map[string]string{"status": "ok"})
}

// @Summary Список моделей AI
// @Description Возвращает список доступных моделей Gemini для пользователя с подробной информацией
// @Tags ai
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.AIModelsResponse
// @Failure 401 {object} domain.ErrorResponse
// @Failure 500 {object} domain.ErrorResponse
// @Failure 429 {object} domain.ErrorResponse
// @Router /user/ai/models [get]
func (h *Handler) AIModels(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		utils.Error(c.Writer, http.StatusUnauthorized, "unauthorized", "no claims")
		return
	}
	usersDB := db.NewUsersProvider(h.db)
	user, err := usersDB.GetUserByTelegramID(claims.TgID)
	if err != nil {
		utils.Error(c.Writer, http.StatusUnauthorized, "unauthorizeds", "user not found")
		return
	}
	models, err := h.ai.ListModels(user.GeminiAPIKey.String)
	if err != nil {
		utils.Error(c.Writer, http.StatusInternalServerError, "ai_error", err.Error())
		return
	}
	utils.Success(c.Writer, map[string]interface{}{"models": models})
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
func (h *Handler) AIText(c *gin.Context) {
	var req domain.AITextRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, http.StatusBadRequest, "bad_request", "invalid body")
		return
	}
	if req.Prompt == "" {
		utils.Error(c.Writer, http.StatusBadRequest, "validation_error", "prompt required")
		return
	}
	// Если модель не указана, используем дефолтную
	if req.Model == "" {
		req.Model = "gemini-2.0-flash-exp"
	}
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		utils.Error(c.Writer, http.StatusUnauthorized, "unauthorized", "no claims")
		return
	}
	users := db.NewUsersProvider(h.db)
	user, err := users.GetUserByTelegramID(claims.TgID)
	if err != nil {
		utils.Error(c.Writer, http.StatusUnauthorized, "unauthorized", "user not found")
		return
	}

	// Для локальных моделей ключ не требуется
	apiKey := user.GeminiAPIKey.String
	if !gemini.IsLocalModel(req.Model) {
		if !user.GeminiAPIKey.Valid || user.GeminiAPIKey.String == "" {
			utils.Error(c.Writer, http.StatusBadRequest, "missing_api_key", "set your Gemini API key first")
			return
		}
	}

	text, err := h.ai.AskText(req.Model, apiKey, req.Prompt)
	if err != nil {
		utils.Error(c.Writer, http.StatusInternalServerError, "ai_error", err.Error())
		return
	}
	utils.Success(c.Writer, map[string]string{"text": text})
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
func (h *Handler) AISetKey(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		utils.Error(c.Writer, http.StatusUnauthorized, "unauthorized", "no claims")
		return
	}
	var req domain.SetKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		utils.Error(c.Writer, http.StatusBadRequest, "bad_request", "invalid body")
		return
	}
	if req.APIKey == "" {
		utils.Error(c.Writer, http.StatusBadRequest, "validation_error", "api_key cannot be empty")
		return
	}
	if len(req.APIKey) < 10 {
		utils.Error(c.Writer, http.StatusBadRequest, "validation_error", "api_key too short")
		return
	}
	users := db.NewUsersProvider(h.db)
	if err := users.SetGeminiAPIKey(claims.TgID, req.APIKey); err != nil {
		utils.Error(c.Writer, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	logger.L.Debug("user set gemini key", "tg_id", claims.TgID)
	utils.Success(c.Writer, map[string]string{"status": "ok"})
}

// @Summary Удалить ключ Gemini
// @Tags ai
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.OptionsSuccessResponse
// @Failure 401 {object} domain.ErrorResponse
// @Router /user/ai/key [delete]
func (h *Handler) AIClearKey(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		utils.Error(c.Writer, http.StatusUnauthorized, "unauthorized", "no claims")
		return
	}
	users := db.NewUsersProvider(h.db)
	if err := users.ClearGeminiAPIKey(claims.TgID); err != nil {
		utils.Error(c.Writer, http.StatusInternalServerError, "db_error", err.Error())
		return
	}
	logger.L.Debug("user cleared gemini key", "tg_id", claims.TgID)
	utils.Success(c.Writer, map[string]string{"status": "ok"})
}

// @Summary Статус ключа Gemini
// @Tags ai
// @Produce json
// @Security BearerAuth
// @Success 200 {object} domain.KeyStatusResponse
// @Failure 401 {object} domain.ErrorResponse
// @Router /user/ai/key [get]
func (h *Handler) AIKeyStatus(c *gin.Context) {
	claims, ok := middleware.ClaimsFromContext(c)
	if !ok {
		utils.Error(c.Writer, http.StatusUnauthorized, "unauthorized", "no claims")
		return
	}
	users := db.NewUsersProvider(h.db)
	user, err := users.GetUserByTelegramID(claims.TgID)
	if err != nil {
		utils.Error(c.Writer, http.StatusUnauthorized, "unauthorized", "user not found")
		return
	}
	utils.Success(c.Writer, domain.KeyStatusResponse{HasKey: user.GeminiAPIKey.Valid && user.GeminiAPIKey.String != ""})
}

func (h *Handler) AdminPing(c *gin.Context) {
	utils.Success(c.Writer, map[string]string{"message": "pong from admin"})
}

func (h *Handler) UserPing(c *gin.Context) {
	utils.Success(c.Writer, map[string]string{"message": "pong from user"})
}

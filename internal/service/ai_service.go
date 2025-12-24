package service

import (
	"geminiBackend/config"
	"geminiBackend/internal/domain"
	"geminiBackend/internal/provider/gemini"
	"strings"
)

type AIService struct {
	cfg *config.Config
}

func NewAIService(cfg *config.Config) *AIService {
	return &AIService{cfg: cfg}
}

func (s *AIService) AskText(model, apiKey, prompt string) (string, error) {
	// Определяем, локальная это модель или облачная по названию
	if isLocalModel(model) {
		localClient := gemini.NewLocalLLMClient(
			s.cfg.LocalLLMEndpoint,
			model,
			s.cfg.LocalLLMMaxChars,
		)
		return localClient.GenerateTextChunked(prompt, s.cfg.LocalLLMMaxChars)
	}

	// Иначе используем Gemini
	client := gemini.NewClient(apiKey, model)
	return client.GenerateText(prompt)
}

// isLocalModel проверяет, является ли модель локальной
func isLocalModel(model string) bool {
	// Список префиксов/имен локальных моделей
	localModels := []string{
		"local",
		"qwen",
		"phi",
		"llama",
		"mistral",
		"gemma",
	}

	modelLower := strings.ToLower(model)
	for _, local := range localModels {
		if strings.HasPrefix(modelLower, local) {
			return true
		}
	}
	return false
}

func (s *AIService) ListModels(apiKey string) ([]domain.ModelInfo, error) {
	client := gemini.NewClient(apiKey, "")
	allModels, err := client.GetAvailableModels()
	if err != nil {
		return nil, err
	}

	// Фильтруем только активные модели
	activeModels := make([]domain.ModelInfo, 0)
	for _, model := range allModels {
		if model.IsAvailable {
			activeModels = append(activeModels, model)
		}
	}

	return activeModels, nil
}

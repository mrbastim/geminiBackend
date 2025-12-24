package service

import (
	"geminiBackend/config"
	"geminiBackend/internal/domain"
	"geminiBackend/internal/provider/gemini"
)

type AIService struct {
	cfg *config.Config
}

func NewAIService(cfg *config.Config) *AIService {
	return &AIService{cfg: cfg}
}

func (s *AIService) AskText(model, apiKey, prompt string) (string, error) {
	// Определяем, локальная это модель или облачная по названию
	if gemini.IsLocalModel(model) {
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

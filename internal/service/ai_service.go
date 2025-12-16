package service

import (
	"geminiBackend/internal/domain"
	"geminiBackend/internal/provider/gemini"
)

type AIService struct{}

func NewAIService() *AIService { return &AIService{} }

func (s *AIService) AskText(model, apiKey, prompt string) (string, error) {
	client := gemini.NewClient(apiKey, model)
	return client.GenerateText(prompt)
}

func (s *AIService) ListModels(apiKey string) ([]domain.ModelInfo, error) {
	client := gemini.NewClient(apiKey, "")
	return client.GetAvailableModels()
}

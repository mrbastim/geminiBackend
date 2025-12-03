package service

import "geminiBackend/internal/provider/gemini"

type AIService struct{}

func NewAIService() *AIService { return &AIService{} }

func (s *AIService) AskText(apiKey, prompt string) (string, error) {
	client := gemini.NewClient(apiKey)
	return client.GenerateText(prompt)
}

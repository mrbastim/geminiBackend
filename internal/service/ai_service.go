package service

import "geminiBackend/internal/provider/gemini"

type AIService struct{ client *gemini.Client }

func NewAIService(client *gemini.Client) *AIService { return &AIService{client: client} }

func (s *AIService) AskText(prompt string) (string, error) { return s.client.GenerateText(prompt) }

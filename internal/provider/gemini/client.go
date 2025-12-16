package gemini

import (
	"context"
	"geminiBackend/internal/domain"
	"strings"

	"google.golang.org/genai"
)

type Client struct {
	apiKey string
	model  string
}

func NewClient(apiKey, model string) *Client { return &Client{apiKey: apiKey, model: model} }

// categorizeModel определяет категорию модели по её имени
func categorizeModel(name string) string {
	lowerName := strings.ToLower(name)

	if strings.Contains(lowerName, "embedding") || strings.Contains(lowerName, "aqa") {
		return "embedding"
	}
	if strings.Contains(lowerName, "imagen") {
		return "image-generation"
	}
	if strings.Contains(lowerName, "veo") {
		return "video-generation"
	}
	if strings.Contains(lowerName, "gemma") {
		return "text"
	}
	if strings.Contains(lowerName, "gemini") {
		// Проверяем на специализированные версии
		if strings.Contains(lowerName, "robotics") {
			return "robotics"
		}
		if strings.Contains(lowerName, "computer-use") {
			return "computer-use"
		}
		if strings.Contains(lowerName, "deep-research") {
			return "research"
		}
		if strings.Contains(lowerName, "tts") || strings.Contains(lowerName, "audio") {
			return "audio"
		}
		if strings.Contains(lowerName, "image") {
			return "image-analysis"
		}
		return "multimodal" // Большинство Gemini моделей
	}

	return "other"
}

func (c *Client) GetAvailableModels() ([]domain.ModelInfo, error) {
	ctx := context.Background()
	models := []domain.ModelInfo{}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  c.apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, err
	}

	for model, err := range client.Models.All(ctx) {
		if err != nil {
			return nil, err
		}

		// Проверяем, поддерживает ли модель генерацию контента
		supportsGeneration := false
		for _, action := range model.SupportedActions {
			if action == "generateContent" {
				supportsGeneration = true
				break
			}
		}

		modelInfo := domain.ModelInfo{
			Name:             model.Name,
			DisplayName:      model.DisplayName,
			Description:      model.Description,
			SupportedActions: model.SupportedActions,
			InputTokenLimit:  model.InputTokenLimit,
			OutputTokenLimit: model.OutputTokenLimit,
			Category:         categorizeModel(model.Name),
			IsAvailable:      supportsGeneration, // Считаем доступной, если поддерживает generateContent
		}

		models = append(models, modelInfo)
	}

	return models, nil
}

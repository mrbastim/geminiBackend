package gemini

import (
	"context"

	"google.golang.org/genai"
)

type Client struct {
	apiKey string
	model  string
}

func NewClient(apiKey, model string) *Client { return &Client{apiKey: apiKey, model: model} }

func (c *Client) GetAvailableModels() ([]string, error) {
	ctx := context.Background()
	models := []string{}

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
		models = append(models, model.Name)
	}

	return models, nil
}

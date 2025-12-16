package gemini

import (
	"context"
	"geminiBackend/pkg/logger"

	"google.golang.org/genai"
)

func (c *Client) GenerateText(prompt string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: c.apiKey})
	if err != nil {
		logger.L.Error("failed to create genai client", "error", err.Error())
		return "", err
	}

	// Используем модель из клиента, если не указана - дефолтная gemini-2.0-flash-exp
	model := c.model
	if model == "" {
		model = "gemini-2.0-flash-exp"
	}

	result, err := client.Models.GenerateContent(
		ctx,
		model,
		genai.Text(prompt),
		&genai.GenerateContentConfig{},
	)
	if err != nil {
		logger.L.Error("failed to generate content", "error", err.Error(), "model", model)
		return "", err
	}
	return result.Text(), nil
}

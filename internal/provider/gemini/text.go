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

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-flash-latest",
		genai.Text(prompt),
		&genai.GenerateContentConfig{},
	)
	if err != nil {
		logger.L.Error("failed to generate content", "error", err.Error())
		return "", err
	}
	return result.Text(), nil
}

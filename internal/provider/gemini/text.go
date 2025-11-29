package gemini

import (
	"context"
	"log"

	"google.golang.org/genai"
)

func (c *Client) GenerateText(prompt string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{APIKey: c.apiKey})
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		&genai.GenerateContentConfig{},
	)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	return result.Text(), nil
}

package gemini

import (
	"context"
	"log"

	"google.golang.org/genai"
)

func (c *Client) GenerateText(prompt string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, nil)
	if err != nil {
		log.Fatal(err)
		return "", err
	}

	result, err := client.Models.GenerateContent(
		ctx,
		"gemini-2.5-flash",
		genai.Text(prompt),
		nil,
	)
	if err != nil {
		log.Fatal(err)
		return "", err
	}
	return result.Text(), nil
}

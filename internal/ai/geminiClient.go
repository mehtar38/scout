package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type Client struct {
	model *genai.GenerativeModel
	ctx   context.Context
}

func NewClient(apiKey string) (*Client, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create AI client: %w", err)
	}

	model := client.GenerativeModel("gemini-2.5-pro")
	model.ResponseMIMEType = "application/json"

	return &Client{
		model: model,
		ctx:   ctx,
	}, nil
}

func (c *Client) ParseQuery(userQuery string) (*CommandParser, error) {
	prompt := BuildPrompt(userQuery)

	resp, err := c.model.GenerateContent(c.ctx, genai.Text(prompt))
	if err != nil {
		return nil, fmt.Errorf("AI request failed: %w", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from AI")
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	var cmd CommandParser
	if err := json.Unmarshal([]byte(responseText), &cmd); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	return &cmd, nil
}

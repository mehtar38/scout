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

	model := client.GenerativeModel("gemini-2.5-flash")

	return &Client{
		model: model,
		ctx:   ctx,
	}, nil
}

func (c *Client) ParseQuery(userQuery string) (*CommandParser, error) {
	c.model.ResponseMIMEType = "application/json"
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

func (c *Client) ProcessDocument(prompt string) (string, error) {
	c.model.ResponseMIMEType = "text/plain" // Set to PLAIN TEXT for document actions
	resp, err := c.model.GenerateContent(c.ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response")
	}
	return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil
}

// func (c *Client) SimpleQuery(prompt string) (string, error) {
// 	ctx := context.Background()

// 	resp, err := c.model.GenerateContent(ctx, genai.Text(prompt))
// 	if err != nil {
// 		return "", fmt.Errorf("failed to generate content: %w", err)
// 	}

// 	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
// 		return "", fmt.Errorf("empty response from AI")
// 	}

// 	// Extract text from response
// 	var result strings.Builder
// 	for _, part := range resp.Candidates[0].Content.Parts {
// 		result.WriteString(fmt.Sprintf("%v", part))
// 	}

// 	return result.String(), nil
// }

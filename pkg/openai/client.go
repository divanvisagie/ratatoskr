package openai

import (
	"context"

	openai "github.com/sashabaranov/go-openai"
)

type Client struct {
	api          *openai.Client
	systemPrompt string
}

func NewClient(apiKey string) *Client {
	client := openai.NewClient(apiKey)

	systemPrompt := "You are Ratatoskr, a telegram bot written in golang, answer users questions in telegram compatible markdown"
	return &Client{api: client, systemPrompt: systemPrompt}
}

func (c *Client) GetCompletion(prompt string) (string, error) {
	req := openai.ChatCompletionRequest{
		Model: "gpt-4o",
		Messages: []openai.ChatCompletionMessage{
			{Role: "system", Content: c.systemPrompt},
			{Role: "user", Content: prompt},
		},
		MaxTokens: 100,
	}
	resp, err := c.api.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

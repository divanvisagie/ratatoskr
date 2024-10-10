package openai

import (
	"context"
	"log"

	openai "github.com/sashabaranov/go-openai"
)

type ChatClient struct {
	api          *openai.Client
	systemPrompt string
}

func NewClient(apiKey string) *ChatClient {
	
	client := openai.NewClient(apiKey)

	systemPrompt := "You are Ratatoskr, a telegram bot written in golang, answer users questions in telegram compatible markdown"
	return &ChatClient{api: client, systemPrompt: systemPrompt}
}

func (c *ChatClient) SetSystemPrompt(prompt string) {
	log.Println("Setting system prompt to: ", prompt)
	c.systemPrompt = prompt
}

func (c *ChatClient) GetCompletion(prompt string) (string, error) {
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

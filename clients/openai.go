package client

import (
	"context"
	"log"
	"os"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAiClient struct {
	// The bot token
	systemPrompt string
	client       *openai.Client
	maxTokens    int
	history      []openai.ChatCompletionMessage
}

func NewOpenAIClient(prompt string) *OpenAiClient {
	token := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(token)
	return &OpenAiClient{
		client:       client,
		maxTokens:    750,
		systemPrompt: prompt,
		history:      []openai.ChatCompletionMessage{},
	}
}

func (c *OpenAiClient) SetMaxTokens(maxTokens int) *OpenAiClient {
	c.maxTokens = maxTokens
	return c
}

func (c *OpenAiClient) SetHistory(history []openai.ChatCompletionMessage) *OpenAiClient {
	c.history = history
	return c
}

func (c *OpenAiClient) Complete(message string) string {

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: c.systemPrompt,
		},
	}

	messages = append(messages, c.history...)

	prmt := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	}
	messages = append(messages, prmt)

	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo,
			Messages:  messages,
			MaxTokens: c.maxTokens,
		},
	)

	if err != nil {
		log.Println(err)
	}

	return resp.Choices[0].Message.Content
}

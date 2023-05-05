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
	context      []openai.ChatCompletionMessage
}

func NewOpenAIClient(prompt string) *OpenAiClient {
	token := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(token)
	return &OpenAiClient{
		client:       client,
		maxTokens:    750,
		systemPrompt: prompt,
		context:      []openai.ChatCompletionMessage{},
	}
}

func (c *OpenAiClient) SetMaxTokens(maxTokens int) *OpenAiClient {
	c.maxTokens = maxTokens
	return c
}

func (c *OpenAiClient) SetHistory(history []openai.ChatCompletionMessage) *OpenAiClient {
	c.context = history
	return c
}

func (c *OpenAiClient) AddUserMessage(message string) *OpenAiClient {
	prmt := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	}
	c.context = append(c.context, prmt)

	return c
}

func (c *OpenAiClient) Complete() string {
	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo,
			Messages:  c.context,
			MaxTokens: c.maxTokens,
		},
	)

	if err != nil {
		log.Println(err)
	}

	if len(resp.Choices) == 0 {
		return "I was unable to summarise this article"
	}

	return resp.Choices[0].Message.Content
}

package client

import (
	"context"
	"log"
	"os"
	"ratatoskr/utils"

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
		maxTokens:    512,
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
	ctx, err := utils.ShortenContext(c.context, utils.MODEL_LIMIT - c.maxTokens)

	if err != nil {
		log.Println(err)
		return `There was an error while trying to shorten the context in the OpenAI client module.`
	}

	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT3Dot5Turbo,
			Messages:  ctx,
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

func Embed(message string) []float32 {
	token := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(token)

	req := openai.EmbeddingRequest{
		Model: openai.AdaEmbeddingV2,
		Input: []string{
			message,
		},
	}

	resp, err := client.CreateEmbeddings(context.Background(), req)
	if err != nil {
		log.Println(err)
	}
	return resp.Data[0].Embedding
}

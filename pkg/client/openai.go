package client

import (
	"context"
	"os"
	"ratatoskr/internal/utils"
	pu "ratatoskr/pkg/utils"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAiClient struct {
	// The bot token
	systemPrompt openai.ChatCompletionMessage
	client       *openai.Client
	maxTokens    int
	context      []openai.ChatCompletionMessage
	logger       *pu.Logger
}

func NewOpenAIClient() *OpenAiClient {
	token := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(token)
	logger := pu.NewLogger("OpenAI Client")
	return &OpenAiClient{
		systemPrompt: openai.ChatCompletionMessage{},
		client:       client,
		maxTokens:    512, //default
		context:      []openai.ChatCompletionMessage{},
		logger:       logger,
	}
}

func (c *OpenAiClient) SetMaxTokens(maxTokens int) *OpenAiClient {
	c.maxTokens = maxTokens
	return c
}

func (c *OpenAiClient) SetHistory(history []openai.ChatCompletionMessage) *OpenAiClient {
	c.context = append(c.context, history...)
	return c
}

func (c *OpenAiClient) AddSystemMessage(message string) *OpenAiClient {
	c.systemPrompt = openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: message,
	}
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
	ts, err := utils.Tokenize(c.systemPrompt.Content)
	if err != nil {
		c.logger.Error("Complete, tokenize system prompt", err)
		return `There was an error while trying to tokenize the system prompt in the OpenAI client module.`
	}

	ctx, err := utils.ShortenContext(c.context, utils.MODEL_LIMIT-c.maxTokens-len(ts))

	if err != nil {
		c.logger.Error("Complete, shorten context", err)
		return `There was an error while trying to shorten the context in the OpenAI client module.`
	}

	messages := []openai.ChatCompletionMessage{}
	messages = append(messages, c.systemPrompt)
	messages = append(messages, ctx...)

	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:     openai.GPT4,
			Messages:  messages,
			MaxTokens: c.maxTokens,
		},
	)

	if err != nil {
		c.logger.Error("Complete, create chat completion", err)
	}
	if len(resp.Choices) == 0 {
		return "I was unable to summarise this article"
	}

	return resp.Choices[0].Message.Content
}

func (c *OpenAiClient) Embed(message string) []float32 {
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
		c.logger.Error("Embed, create embeddings", err)
	}
	return resp.Data[0].Embedding
}

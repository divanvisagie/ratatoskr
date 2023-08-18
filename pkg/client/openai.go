package client

import (
	"context"
	"log"
	"os"
	"ratatoskr/internal/utils"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAiClient struct {
	// The bot token
	systemPrompt openai.ChatCompletionMessage
	client       *openai.Client
	maxTokens    int
	context      []openai.ChatCompletionMessage
}

func NewOpenAIClient() *OpenAiClient {
	token := os.Getenv("OPENAI_API_KEY")
	client := openai.NewClient(token)
	return &OpenAiClient{
		systemPrompt: openai.ChatCompletionMessage{},
		client:       client,
		maxTokens:    512, //default
		context:      []openai.ChatCompletionMessage{},
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
		log.Printf("Error while tokenizing system prompt: %v", err)
		return `There was an error while trying to tokenize the system prompt in the OpenAI client module.`
	}

	ctx, err := utils.ShortenContext(c.context, utils.MODEL_LIMIT-c.maxTokens-len(ts))

	if err != nil {
		log.Println(err)
		return `There was an error while trying to shorten the context in the OpenAI client module.`
	}

	messages := []openai.ChatCompletionMessage{}
	messages = append(messages, c.systemPrompt)
	messages = append(messages, ctx...)

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

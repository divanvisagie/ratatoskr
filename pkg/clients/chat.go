package clients

import (
	"context"
	"log"

	"github.com/divanvisagie/ratatoskr/pkg/types"
	openai "github.com/sashabaranov/go-openai"
)

type ChatClient struct {
	api     *openai.Client
	context *[]openai.ChatCompletionMessage
}

func NewChatClient(apiKey string) *ChatClient {

	client := openai.NewClient(apiKey)

	context := []openai.ChatCompletionMessage{}
	return &ChatClient{api: client, context: &context}
}

func (c *ChatClient) SetSystemPrompt(prompt string) {
	c.context = &[]openai.ChatCompletionMessage{
		{Role: "system", Content: prompt},
	}
}

func (c *ChatClient) AddMessage(role, content string) {
	*c.context = append(*c.context, openai.ChatCompletionMessage{Role: role, Content: content})
}

func (c *ChatClient) AddStoredMessages(messages []types.StoredMessage) {
	for _, message := range messages {
		c.AddMessage(message.Role, message.Content)
	}
}

func (c *ChatClient) GetCompletion() (string, error) {
	req := openai.ChatCompletionRequest{
		Model:     "gpt-4o",
		Messages:  *c.context,
		MaxTokens: 400,
	}
	log.Println(req)

	resp, err := c.api.CreateChatCompletion(context.Background(), req)
	if err != nil {
		return "", err
	}
	return resp.Choices[0].Message.Content, nil
}

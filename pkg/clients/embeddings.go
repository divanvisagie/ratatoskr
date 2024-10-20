/*
Client for getting openai embeddings
*/
package clients

import (
	"context"

	"github.com/divanvisagie/ratatoskr/internal/logger"
	openai "github.com/sashabaranov/go-openai"
)

type EmbeddingsClient interface {
	GetEmbeddings(text string) ([]float32, error)
}

type OpenAiEmbeddingsClient struct {
	api    *openai.Client
	logger logger.Logger
}

func NewEmbeddingsClient(apiKey string) *OpenAiEmbeddingsClient {
	client := openai.NewClient(apiKey)
	return &OpenAiEmbeddingsClient{api: client, logger: *logger.NewLogger("EmbeddingsClient")}
}

// Get embeddings for string
func (c *OpenAiEmbeddingsClient) GetEmbeddings(text string) ([]float32, error) {
	req := openai.EmbeddingRequest{
		Model: "text-embedding-3-small",
		Input: text,
	}

	resp, err := c.api.CreateEmbeddings(context.Background(), req)
	if err != nil {
		return make([]float32, 0), err
	}

	return resp.Data[0].Embedding, nil
}

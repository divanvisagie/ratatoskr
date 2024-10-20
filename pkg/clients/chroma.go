package clients

import (
	"context"

	ch "github.com/amikos-tech/chroma-go"
	"github.com/amikos-tech/chroma-go/openai"
	ty "github.com/amikos-tech/chroma-go/types"
	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/internal/logger"
)

type ChromaClient struct {
	logger    logger.Logger
	chromaURL string
	cfg       config.Config
	client    *ch.Client
	openaiEf  *openai.OpenAIEmbeddingFunction
}

func NewChromaClient(config config.Config) *ChromaClient {
	logger := *logger.NewLogger("ChromaClient")
	client, err := ch.NewClient(config.ChromaBaseUrl)
	if err != nil {
		logger.Error("Failed to create Chroma client", err)
		panic(err)
	}

	openaiEf, err := openai.NewOpenAIEmbeddingFunction(config.OpenAIKey)
	if err != nil {
		logger.Error("Failed to create OpenAI embedding function", err)
	}

	// Create a new collection with OpenAI embedding function, L2 distance function and metadata
	_, err = client.CreateCollection(context.Background(), "chat_messages", map[string]interface{}{"id": "mc"}, true, openaiEf, ty.L2)
	if err != nil {
		logger.Error("Failed to create Chroma collection", err)
	}

	logger.Info("Chroma client created successfully", client)

	return &ChromaClient{
		chromaURL: config.ChromaBaseUrl,
		cfg:       config,
		logger:    logger,
		client:    client,
		openaiEf:  openaiEf,
	}
}

// SaveEmbeddedVector stores an embedded vector with metadata in Chroma
func (c *ChromaClient) SaveEmbeddedVector(messageId int64, chatId int64, content string) error {
	// Get collection
	collection, err := c.client.GetCollection(context.Background(), "chat_messages", c.openaiEf)
	if err != nil {
		c.logger.Error("Failed to get collection", err)
		return err
	}

	_, err = collection.Add(context.TODO(), nil, []map[string]interface{}{{"id": messageId}}, []string{content}, []string{string(messageId)})
	if err != nil {
		c.logger.Error("Failed to add document", err)
		return err
	}
	return nil
}

// SearchForMessage performs a vector-based search for related messages
func (c *ChromaClient) SearchForMessage(message string, topK int32) ([]int64, error) {
	// Get collection
	collection, err := c.client.GetCollection(context.Background(), "chat_messages", c.openaiEf)
	if err != nil {
		c.logger.Error("Failed to get collection", err)
		return nil, err
	}

	// Perform search

	data, err := collection.Query(context.Background(), []string{message}, topK, nil, nil, nil)
	if err != nil {
		c.logger.Error("Failed to query collection", err)
		return nil, err
	}


	c.logger.Info("Search results", data.Metadatas)

	ids := make([]int64, len(data.Metadatas))
	for i, metadata := range data.Metadatas {
		raw := metadata[0]
		id := raw["id"]
		ids[i] = int64(id.(float64))
	}

	c.logger.Info("Search results", ids)

	return ids, nil
}

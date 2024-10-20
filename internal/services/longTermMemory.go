package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/pkg/clients"
	"github.com/divanvisagie/ratatoskr/pkg/store"
	"github.com/divanvisagie/ratatoskr/pkg/types"
)

type LongTermMemoryService struct {
	store            store.DocumentStore
	embeddingsClient clients.EmbeddingsClient
	logger           logger.Logger
	chromaURL        string
}

func NewLongTermMemoryService(client clients.EmbeddingsClient, cfg config.Config) *LongTermMemoryService {
	return &LongTermMemoryService{
		store:            *store.NewDocumentStore(),
		embeddingsClient: client,
		logger:           *logger.NewLogger("LongTermMemoryService"),
		chromaURL:        cfg.ChromaBaseUrl,
	}
}

// StoreMessageLongTerm generates embeddings and stores them in Chroma
func (l *LongTermMemoryService) StoreMessageLongTerm(id int64, message types.StoredMessage) {
	// Generate vector for message
	result, err := l.embeddingsClient.GetEmbeddings(message.Content)
	if err != nil {
		// Handle error
		l.logger.Error("Failed to generate embeddings for message", err)
		return
	}

	// Prepare payload for Chroma
	payload := map[string]interface{}{
		"vector": result, // The vector you got from the embeddings client
		"metadata": map[string]interface{}{ // Optional: Store message metadata
			"id":      id,
			"content": message.Content,
		},
	}

	// Convert payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		l.logger.Error("Failed to marshal embedding payload", err)
		return
	}

	// Create a request to Chroma
	resp, err := http.Post(fmt.Sprintf("%s/api/v1/embeddings", l.chromaURL), "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		l.logger.Error("Failed to store embedding in Chroma", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		l.logger.Error(fmt.Sprintf("Chroma responded with status: %d", resp.StatusCode))
		return
	}

	l.logger.Info("Successfully stored embedding in Chroma")
}

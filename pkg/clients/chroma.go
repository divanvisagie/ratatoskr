package clients

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/internal/logger"
)

type ChromaClient struct {
	logger    logger.Logger
	chromaURL string
	cfg       config.Config
}

func NewChromaClient(config config.Config) *ChromaClient {
	return &ChromaClient{
		logger:    *logger.NewLogger("ChromaClient"),
		chromaURL: config.ChromaBaseUrl,
		cfg:       config,
	}
}

// SearchForMessage performs a vector-based search for related messages
func (c *ChromaClient) SearchForMessage(vector []float32, topK int) ([]int64, error) {
	// Prepare the query payload
	queryPayload := map[string]interface{}{
		"vector": vector, // The query vector to search for
		"top_k":  topK,   // Number of top related results to retrieve
	}

	// Convert the payload to JSON
	queryBytes, err := json.Marshal(queryPayload)
	if err != nil {
		c.logger.Error("Failed to marshal query payload", err)
		return nil, err
	}

	// Send the POST request to Chroma's /query endpoint
	resp, err := http.Post(fmt.Sprintf("%s/api/v1/query", c.chromaURL), "application/json", bytes.NewBuffer(queryBytes))
	if err != nil {
		c.logger.Error("Failed to query Chroma for related messages", err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error(fmt.Sprintf("Chroma responded with status: %d", resp.StatusCode))
		return nil, fmt.Errorf("Chroma query failed with status: %d", resp.StatusCode)
	}

	// Decode the response into the expected structure
	var searchResults struct {
		Results []struct {
			Score    float64                `json:"score"`
			Metadata map[string]interface{} `json:"metadata"`
		} `json:"results"`
	}

	err = json.NewDecoder(resp.Body).Decode(&searchResults)
	if err != nil {
		c.logger.Error("Failed to decode Chroma search response", err)
		return nil, err
	}

	// Extract message IDs from the results
	var messageIDs []int64
	for _, result := range searchResults.Results {
		if id, ok := result.Metadata["id"].(float64); ok {
			messageIDs = append(messageIDs, int64(id))
		}
	}

	return messageIDs, nil
}

// SaveEmbeddedVector stores an embedded vector with metadata in Chroma
func (c *ChromaClient) SaveEmbeddedVector(vector []float32, content string, messageId int64, chatId int64) error {
	// Prepare payload for Chroma
	payload := map[string]interface{}{
		"vector": vector, // The vector you got from the embeddings client
		"metadata": map[string]interface{}{ // Optional: Store message metadata
			"id":      messageId,
			"chatId":  chatId,
			"content": content,
		},
	}

	// Convert payload to JSON
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		c.logger.Error("Failed to marshal embedding payload", err)
		return nil
	}

	// Send POST request to Chroma's /embeddings endpoint
	resp, err := http.Post(fmt.Sprintf("%s/api/v1/embeddings", c.chromaURL), "application/json", bytes.NewBuffer(payloadBytes))
	if err != nil {
		c.logger.Error("Failed to store embedding in Chroma", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.Error(fmt.Sprintf("Chroma responded with status: %d", resp.StatusCode))
		return nil
	}
	
	return nil
}

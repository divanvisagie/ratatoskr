package repos

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"ratatoskr/client"
	"ratatoskr/types"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
)

type Role string

const (
	System    Role = "system"
	User      Role = "user"
	Assistant Role = "assistant"
)

type Message struct {
	client *redis.Client
}

func NewMessageRepository() *Message {
	url := os.Getenv("REDIS_URL")
	opt, err := redis.ParseURL(url)
	if err != nil {
		log.Printf("Redis is not working: %v", err)
		return nil
	}

	rdb := redis.NewClient(opt)
	return &Message{
		client: rdb,
	}
}

func (m *Message) GetMessages(username string) ([]types.StoredMessage, error) {
	ctx := context.Background()
	// retrieve messages from Redis
	val, err := m.client.Get(ctx, username).Bytes()
	if err != nil {
		log.Printf("Failed to retreive messages from Redis: %v", err)
		return []types.StoredMessage{}, nil
	}

	// unmarshal the JSON-encoded messages
	var messages []types.StoredMessage
	err = json.Unmarshal(val, &messages)
	if err != nil {
		log.Printf("Failed to unmarshal messages: %v", err)
		return nil, err
	}

	// sort and return the messages
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp < messages[j].Timestamp
	})

	return messages, nil
}

func (m *Message) SaveMessage(role Role, username string, message string) error {
	ctx := context.Background()
	// create a new StoredMessage struct
	now := time.Now()
	timestamp := now.UnixMilli()
	storedMessage := types.StoredMessage{
		Role:      string(role),
		Message:   message,
		Timestamp: timestamp,
	}

	// retrieve the existing messages from Redis
	val, err := m.client.Get(ctx, username).Bytes()
	if err != nil && err != redis.Nil {
		log.Printf("Failed to retreive messages from Redis: %v", err)
		return err
	}

	var messages []types.StoredMessage

	// unmarshal the JSON-encoded messages
	if len(val) > 0 {
		err = json.Unmarshal(val, &messages)
		if err != nil {
			log.Printf("Failed to unmarshal messages while saving new message")
			return err
		}
	}

	// append the new message and marshal all messages to JSON
	messages = append(messages, storedMessage)
	if len(messages) > 15 {
		messages = messages[len(messages)-15:]
	}
	jsonBytes, err := json.Marshal(messages)
	if err != nil {
		log.Printf("Failed to marshal messages while saving new message: %v", err)
		return err
	}

	// store the messages in Redis
	err = m.client.Set(ctx, username, jsonBytes, 0).Err()
	if err != nil {
		log.Printf("Failed to save messages to Redis: %v", err)
		return err
	}

	fmt.Printf("Saved message %s from %s at %d\n", message, username, timestamp)

	return nil
}

func (m *Message) ClearMemory() {
	// delete all keys in the Redis database
	ctx := context.Background()
	m.client.FlushAll(ctx)
}

// -----------------------------------------------------------------------------
type Embedding struct {
	Text      string
	Embedding []float32
	Role      Role
	User      string
}

func (m *Message) RememberEmbedded(role Role, username string, message string) {
	// Create openai vector with message
	vector := client.Embed(message)
	println(vector)

	hash := md5.Sum([]byte(message))
	key := fmt.Sprintf("embedding:%x", hash)

	// Save embedding to Redis
	ctx := context.Background()

	embeddingJSON, err := json.Marshal(vector)
	if err != nil {
		log.Printf("Failed to marshal embedding: %v", err)
		return
	}

	m.client.HSet(ctx, key, "embedding", string(embeddingJSON))
	m.client.HSet(ctx, key, "role", string(role))
	m.client.HSet(ctx, key, "user", username)
	m.client.HSet(ctx, key, "text", message)
}

func NewStoredMessage(role Role, message string) *types.StoredMessage {
	now := time.Now()
	timestamp := now.UnixMilli()
	storedMessage := &types.StoredMessage{
		Role:      string(role),
		Message:   message,
		Timestamp: timestamp,
	}
	return storedMessage
}

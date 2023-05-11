package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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
		fmt.Println("Redis is not working")
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
		fmt.Println("Failed to retreive messages from Redis")
		return nil, err
	}

	// unmarshal the JSON-encoded messages
	var messages []types.StoredMessage
	err = json.Unmarshal(val, &messages)
	if err != nil {
		fmt.Println("Failed to unmarshal messages")
		return nil, err
	}

	// sort and return the messages
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp < messages[j].Timestamp
	})

	if len(messages) > 20 {
		messages = messages[len(messages)-20:]
	}

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
		return err
	}

	var messages []types.StoredMessage

	// unmarshal the JSON-encoded messages
	if len(val) > 0 {
		err = json.Unmarshal(val, &messages)
		if err != nil {
			fmt.Println("Failed to unmarshal messages while saving new message")
			return err
		}
	}

	// append the new message and marshal all messages to JSON
	messages = append(messages, storedMessage)
	if len(messages) > 20 {
		messages = messages[len(messages)-20:]
	}
	jsonBytes, err := json.Marshal(messages)
	if err != nil {
		fmt.Println("Failed to marshal messages while saving new message")
		return err
	}

	// store the messages in Redis
	err = m.client.Set(ctx, username, jsonBytes, 0).Err()
	if err != nil {
		fmt.Println("Failed to save messages to Redis")
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

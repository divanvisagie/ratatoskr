package repos

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"ratatoskr/pkg/client"
	"ratatoskr/pkg/types"
	"io/ioutil"
	"sort"
	"time"

	"github.com/redis/go-redis/v9"
	"gopkg.in/yaml.v3"
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

// Save the message to a file in the folder structure /data/{year}/{month}/{day}
// the directory is created if it does not exist
func saveMessageInYaml(message types.StoredMessage) {
	root := os.Getenv("ROOT_DIR")

	now := time.Now()
	year := now.Year()
	month := now.Month()
	day := now.Day()

	path := fmt.Sprintf("%s/%d/%d/%d", root, year, month, day)
	log.Printf("Saving message to %s", path)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		log.Printf("Failed to create directory: %v", err)
		return
	}

	filename := fmt.Sprintf("%s/messsages.yaml", path)

	// Read existing data from the file, if it exists
	var storedMessages []types.MessageOnDisk
	if _, err := os.Stat(filename); err == nil {
		fileContents, err := ioutil.ReadFile(filename)
		if err != nil {
			log.Printf("Failed to read file contents: %v", err)
			return
		}

		err = yaml.Unmarshal(fileContents, &storedMessages)
		if err != nil {
			log.Printf("Failed to unmarshal yaml: %v", err)
			return
		}
	}

	// Append new message to the array
	mod := types.MessageOnDisk{
		Role:    message.Role,
		Content: message.Message,
	}
	storedMessages = append(storedMessages, mod)

	// Convert storedMessage array to yaml
	yamlBytes, err := yaml.Marshal(storedMessages)
	if err != nil {
		log.Printf("Failed to marshal yaml: %v", err)
		return
	}

	// Write yaml to file
	err = ioutil.WriteFile(filename, yamlBytes, 0644)
	if err != nil {
		log.Printf("Failed to write yaml to file: %v", err)
		return
	}
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

	saveMessageInYaml(storedMessage)

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

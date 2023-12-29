package repos

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"ratatoskr/pkg/types"
	"time"

	"gopkg.in/yaml.v3"
)

type Role string

const (
	System    Role = "system"
	User      Role = "user"
	Assistant Role = "assistant"
)

type EmbeddingOnDisk struct {
	HashedMessage string
	Embedding []float32
}

type MessageRepo struct {
}

func NewMessageRepository() *MessageRepo {

	return &MessageRepo{}
}

func (m *MessageRepo) GetMessages(username string) ([]types.StoredMessage, error) {
	root := os.Getenv("ROOT_DIR")

	now := time.Now()
	year := now.Year()
	month := now.Month()
	day := now.Day()

	// Get the messages from the yaml file if it exists
	path := fmt.Sprintf("%s/%s/%d/%d/%d/messages.yaml", root, username, year, month, day)

	var messagesOnDisk []types.MessageOnDisk
	if _, err := os.Stat(path); err == nil {
		// Read existing data from the file
		fileContents, err := ioutil.ReadFile(path)
		if err != nil {
			log.Printf("Failed to read file contents: %v", err)
			return nil, err
		}
		err = yaml.Unmarshal(fileContents, &messagesOnDisk)
		if err != nil {
			log.Printf("Failed to unmarshal yaml: %v", err)
			return nil, err
		}

	}

	var storedMessages = make([]types.StoredMessage, len(messagesOnDisk))
	for i, message := range messagesOnDisk {
		storedMessages[i] = types.StoredMessage{
			Role:      message.Role,
			Message:   message.Content,
			Timestamp: 0,
		}
	}


	return storedMessages, nil
}

// Save the message to a file in the folder structure /data/{year}/{month}/{day}
// the directory is created if it does not exist
func saveMessageInYaml(username string, message types.StoredMessage) error {
	root := os.Getenv("ROOT_DIR")

	now := time.Now()
	year := now.Year()
	month := now.Month()
	day := now.Day()

	path := fmt.Sprintf("%s/%s/%d/%d/%d", root, username, year, month, day)
	log.Printf("Saving message to %s", path)
	err := os.MkdirAll(path, 0755)
	if err != nil {
		log.Printf("Failed to create directory: %v", err)
		return err
	}

	msgFile := fmt.Sprintf("%s/messages.yaml", path)

	// Read existing data from the file, if it exists
	var storedMessages []types.MessageOnDisk
	if _, err := os.Stat(msgFile); err == nil {
		fileContents, err := ioutil.ReadFile(msgFile)
		if err != nil {
			log.Printf("Failed to read file contents: %v", err)
			return err
		}

		err = yaml.Unmarshal(fileContents, &storedMessages)
		if err != nil {
			log.Printf("Failed to unmarshal yaml: %v", err)
			return err
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
		return err
	}

	// Write yaml to file
	err = ioutil.WriteFile(msgFile, yamlBytes, 0644)
	if err != nil {
		log.Printf("Failed to write yaml to file: %v", err)
		return err
	}

	return nil
}

func (m *MessageRepo) SaveMessage(username string, message types.StoredMessage) error {
	err := saveMessageInYaml(username, message) 
	return err
}

func (m *MessageRepo) ClearMemory() {
	fmt.Println("Asked for memory clear but functionality is not implemented")
}

// -----------------------------------------------------------------------------
type Embedding struct {
	Text      string
	Embedding []float32
	Role      Role
	User      string
}

func (m *MessageRepo) RememberEmbedded(role Role, username string, message string) {
	fmt.Println("Asked to remember embedded but function is not implemented")
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

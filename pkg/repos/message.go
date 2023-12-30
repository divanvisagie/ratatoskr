package repos

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"ratatoskr/pkg/client"
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
	Hash      string    `yaml:"hash"`
	Embedding []float32 `yaml:"embedding"`
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

func readMessagesFromDisk(msgFile string) ([]types.MessageOnDisk, error) {
	var storedMessages []types.MessageOnDisk
	if _, err := os.Stat(msgFile); err == nil {
		fileContents, err := ioutil.ReadFile(msgFile)
		if err != nil {
			log.Printf("Failed to read file contents: %v", err)
			return nil, err
		}

		err = yaml.Unmarshal(fileContents, &storedMessages)
		if err != nil {
			log.Printf("Failed to unmarshal yaml: %v", err)
			return nil, err
		}
	}
	return storedMessages, nil
}

func readEmbeddingsFromDisk(filePath string) ([]EmbeddingOnDisk, error) {
	var storedMessages []EmbeddingOnDisk
	if _, err := os.Stat(filePath); err == nil {
		fileContents, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Printf("Failed to read file contents: %v", err)
			return nil, err
		}

		err = yaml.Unmarshal(fileContents, &storedMessages)
		if err != nil {
			log.Printf("Failed to unmarshal yaml: %v", err)
			return nil, err
		}
	}
	return storedMessages, nil
}

func getHashOfString(message string) string {
	hasher := md5.New()
	io.WriteString(hasher, message)
	hash := fmt.Sprintf("%x", hasher.Sum(nil))
	return hash
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
	embeddingsFile := fmt.Sprintf("%s/embeddings.yaml", path)

	storedMessages, err := readMessagesFromDisk(msgFile)
	if err != nil {
		log.Printf("Failed to read messages from disk: %v", err)
		return err
	}

	embeddings, err := readEmbeddingsFromDisk(embeddingsFile)
	if err != nil {
		log.Printf("Failed to read embeddings from disk: %v", err)
		return err
	}

	// Append new message to the array
	newMessage := types.MessageOnDisk{
		Role:    message.Role,
		Content: message.Message,
		Hash:    getHashOfString(message.Message),
	}
	storedMessages = append(storedMessages, newMessage)

	//  Create embedding
	embedding := client.Embed(message.Message)
	newEmbedding := EmbeddingOnDisk{
		Hash:      getHashOfString(message.Message),
		Embedding: embedding,
	}
	embeddings = append(embeddings, newEmbedding)

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

	eyamlBytes, err := yaml.Marshal(embeddings)
	if err != nil {
		log.Printf("Failed to marshal yaml: %v", err)
		return err
	}
	err = ioutil.WriteFile(embeddingsFile, eyamlBytes, 0644)
	if err != nil {
		log.Printf("Failed to write embeddings to file: %v", err)
		return err
	}

	return nil
}

func findInstancesOfInPath(path, filename string) (error, []string) {
	pathList := []string{}
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info.Name() == filename {
			pathList = append(pathList, path)
		}
		return nil
	})
	if err != nil {
		return err, nil
	}
	return nil, pathList
}

func (m *MessageRepo) GetAllMessagesForUser(username string) (error, []types.StoredMessage) {
	root := os.Getenv("ROOT_DIR")
	path := fmt.Sprintf("%s/%s", root, username)

	// Search all subdirectories to find paths to all instances of messages.yaml
	err, pathList := findInstancesOfInPath(path, "messages.yaml")
	if err != nil {
		fmt.Println("Failed to find instances of messages.yaml")
		return err, nil
	}

	diskMsgs := []types.MessageOnDisk{}
	for _, path := range pathList {
		messages, err := readMessagesFromDisk(path)
		if err != nil {
			return err, nil
		}
		diskMsgs = append(diskMsgs, messages...)
	}

	// lets get the embeddings now
	err, embeddingPathList := findInstancesOfInPath(path, "embeddings.yaml")
	if err != nil {
		fmt.Println("Failed to find instances of embeddings.yaml")
		return err, nil
	}

	// create a hashmap of embeddings and hashes
	embeddings := make(map[string][]float32)
	for _, path := range embeddingPathList {
		embeddingsOnDisk, err := readEmbeddingsFromDisk(path)
		if err != nil {
			return err, nil
		}
		for _, embedding := range embeddingsOnDisk {
			embeddings[embedding.Hash] = embedding.Embedding
		}
	}

	storedMessages := []types.StoredMessage{}
	for _, msg := range diskMsgs {
		// check if embedding exists for this hash
		if _, ok := embeddings[msg.Hash]; !ok {
			fmt.Printf("No embedding found for hash %s\n", msg.Hash)
			continue
		}
		ebd := embeddings[msg.Hash]
		storedMessages = append(storedMessages, types.StoredMessage{
			Role:      msg.Role,
			Message:   msg.Content,
			Embedding: ebd,
		})
	}
	fmt.Printf("Returning %d messages\n", len(storedMessages))
	return nil, storedMessages
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

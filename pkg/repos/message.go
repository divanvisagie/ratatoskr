package repos

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"ratatoskr/pkg/client"
	"ratatoskr/pkg/types"
	"ratatoskr/pkg/utils"
	"strings"
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
	Hash      string    `json:"hash"`
	Embedding []float32 `json:"embedding"`
}

type EmbeddingInMemory struct {
	Hash      string
	Embedding []float32
	Filename  string
}

type MessageInMemory struct {
	Role     string
	Content  string
	Hash     string
	Filename string
}

type MessageRepo struct {
	logger *utils.Logger
}

func NewMessageRepository() *MessageRepo {
	logger := utils.NewLogger("Message Repository")
	return &MessageRepo{logger}
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
			m.logger.Error("GetMessages, unmarshal yaml", err)
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
			return nil, err
		}

		err = yaml.Unmarshal(fileContents, &storedMessages)
		if err != nil {
			return nil, err
		}
	}

	missingHashes := 0
	for i := 0; i < len(storedMessages); i++ {
		if storedMessages[i].Hash == "" {
			missingHashes++
			storedMessages[i].Hash = getHashOfString(storedMessages[i].Content)
		}
	}

	//if there were missing hashes we should save the state of the new ones
	if missingHashes > 0 {
		yamlBytes, err := yaml.Marshal(storedMessages)
		if err != nil {
			return nil, err
		}
		err = ioutil.WriteFile(msgFile, yamlBytes, 0644)
		if err != nil {
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
			return nil, err
		}

		err = json.Unmarshal(fileContents, &storedMessages)
		if err != nil {
			return nil, err
		}
	}
	return storedMessages, nil
}

func getHashOfString(message string) string {
	hasher := sha256.New()
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
	embeddingsFile := fmt.Sprintf("%s/embeddings.json", path)

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
	c := client.NewOpenAIClient()
	embedding := c.Embed(message.Message)
	newEmbedding := EmbeddingOnDisk{
		Hash:      getHashOfString(message.Message),
		Embedding: embedding,
	}
	embeddings = append(embeddings, newEmbedding)

	// Convert storedMessage array to yaml
	messageJsonBytes, err := yaml.Marshal(storedMessages)
	if err != nil {
		log.Printf("Failed to marshal yaml: %v", err)
		return err
	}

	// Write yaml to file
	err = ioutil.WriteFile(msgFile, messageJsonBytes, 0644)
	if err != nil {
		log.Printf("Failed to write yaml to file: %v", err)
		return err
	}

	embeddingJsonBytes, err := json.Marshal(embeddings)
	if err != nil {
		log.Printf("Failed to marshal json: %v", err)
		return err
	}
	err = ioutil.WriteFile(embeddingsFile, embeddingJsonBytes, 0644)
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

func getEmbeddingsMapWithPaths(pathList []string) (map[string]EmbeddingInMemory, error) {
	embeddings := make(map[string]EmbeddingInMemory)
	for _, path := range pathList {
		embeddingsOnDisk, err := readEmbeddingsFromDisk(path)
		if err != nil {
			return nil, err
		}
		for i := 0; i < len(embeddingsOnDisk); i++ {
			embeddings[embeddingsOnDisk[i].Hash] = EmbeddingInMemory{
				Hash:      embeddingsOnDisk[i].Hash,
				Embedding: embeddingsOnDisk[i].Embedding,
				Filename:  path,
			}
		}
	}

	for hash, embedding := range embeddings {
		fmt.Printf(">> getEmbeddingsMapWithPaths Hash: %s, Filename: %s\n", hash, embedding.Filename)
	}
	return embeddings, nil
}

func (m *MessageRepo) GetAllMessagesForUser(username string) (error, []types.StoredMessage) {
	root := os.Getenv("ROOT_DIR")
	path := fmt.Sprintf("%s/%s", root, username)

	// Search all subdirectories to find paths to all instances of messages.yaml
	err, pathList := findInstancesOfInPath(path, "messages.yaml")
	if err != nil {
		m.logger.Error("GetAllMessagesForUser, findInstancesOfInPath", err)
		return err, nil
	}

	// Get list of messages.yaml files and populate diskMsgs
	diskMsgs := []MessageInMemory{}
	for _, path := range pathList {
		messages, err := readMessagesFromDisk(path)
		if err != nil {
			return err, nil
		}
		//diskMsgs = append(diskMsgs, messages...)
		for _, msg := range messages {
			diskMsgs = append(diskMsgs, MessageInMemory{
				Role:     msg.Role,
				Content:  msg.Content,
				Hash:     msg.Hash,
				Filename: path,
			})
		}
	}

	// lets get the embeddings now
	err, embeddingPathList := findInstancesOfInPath(path, "embeddings.json")
	if err != nil {
		m.logger.Error("GetAllMessagesForUser, findInstancesOfInPath", err)
		return err, nil
	}

	// create a hashmap of embeddings and hashes
	embeddings, err := getEmbeddingsMapWithPaths(embeddingPathList)
	if err != nil {
		m.logger.Error("GetAllMessagesForUser, getEmbeddingsMapWithPaths", err)
		return err, nil
	}

	// Add embeddings to the stored messages
	var newEmbeddingsToBeStored = make(map[string][]EmbeddingOnDisk)
	storedMessages := []types.StoredMessage{}
	for i := 0; i < len(diskMsgs); i++ {
		// check if embedding exists for this hash
		if _, ok := embeddings[diskMsgs[i].Hash]; !ok {
			m.logger.Info("GetAllMessagesForUser, no embedding found for hash %s\n", diskMsgs[i].Hash)
			c := client.NewOpenAIClient()
			embedding := c.Embed(diskMsgs[i].Content)
			// remove the messages.yaml and replace it with embeddings.json
			filename := strings.Replace(diskMsgs[i].Filename, "messages.yaml", "embeddings.json", 1)
			m.logger.Info("GetAllMessagesForUser, setting file path for new embedding to %s\n", filename)
			embeddings[diskMsgs[i].Hash] = EmbeddingInMemory{
				Hash:      diskMsgs[i].Hash,
				Embedding: embedding,
				Filename:  filename,
			}

			towrite := EmbeddingOnDisk{
				Hash:      diskMsgs[i].Hash,
				Embedding: embedding,
			}

			// append embeddingondisk
			if _, ok := newEmbeddingsToBeStored[filename]; !ok {
				newEmbeddingsToBeStored[filename] = []EmbeddingOnDisk{}
			}
			newEmbeddingsToBeStored[filename] = append(newEmbeddingsToBeStored[filename], towrite)

		}
		ebd := embeddings[diskMsgs[i].Hash]
		storedMessages = append(storedMessages, types.StoredMessage{
			Role:      diskMsgs[i].Role,
			Message:   diskMsgs[i].Content,
			Embedding: ebd.Embedding,
		})
	}

	// write new embeddings
	for hash, embeddings := range newEmbeddingsToBeStored {
		m.logger.Info(">> GetAllMessagesForUser, writing new embeddings for hash %s\n", hash)
		existing, err := readEmbeddingsFromDisk(hash)
		if err != nil {
			return err, nil
		}
		existing = append(existing, embeddings...)

		// write to file
		embeddingJsonBytes, err := json.Marshal(embeddings)
		if err != nil {
			m.logger.Error("GetAllMessagesForUser, marshal json", err)
			return err, nil
		}
		err = ioutil.WriteFile(hash, embeddingJsonBytes, 0644)
		if err != nil {
			m.logger.Error("GetAllMessagesForUser, write file", err)
			return err, nil
		}
	}

	m.logger.Info("Returning %d messages\n", len(storedMessages))
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

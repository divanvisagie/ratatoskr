package layers

import (
	"fmt"
	"time"

	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/pkg/store"
	"github.com/divanvisagie/ratatoskr/pkg/types"
)

type MemoryLayer struct {
	out    chan types.ResponseMessage
	next   types.Cortex
	logger *logger.Logger
	store  *store.DocumentStore
}

func (m *MemoryLayer) listenAndRespond() {
	for response := range m.next.GetUpdatesChan() {
		storedMessage := types.StoredMessage{
			Content:   response.Message,
			Role:      "assistant",
			CreatedAt: time.Now().Unix(),
		}
		partitionKey := fmt.Sprintf("user#%d", response.UserId)
		sortKey := fmt.Sprintf("message#%d#%d", response.ChatId, storedMessage.CreatedAt)
		m.store.SaveItem(partitionKey, sortKey, storedMessage)

		m.out <- response
	}
}

func NewMemoryLayer(nextLayer types.Cortex) *MemoryLayer {
	instance := &MemoryLayer{
		out:    make(chan types.ResponseMessage),
		next:   nextLayer,
		logger: logger.NewLogger("MemoryLayer"),
		store:  store.NewDocumentStore(),
	}
	go instance.listenAndRespond()
	return instance
}

func (m *MemoryLayer) SendMessage(message types.RequestMessage) {
	now := time.Now().Unix()
	storedMessage := types.StoredMessage{
		Content:   message.Message,
		Role:      "user",
		CreatedAt: now,
	}

	partitionKey := fmt.Sprintf("user#%d", message.UserId)
	sortKey := fmt.Sprintf("message#%d#%d", message.ChatId, now)

	history, err := m.store.GetItems(partitionKey, fmt.Sprintf("message#%d", message.ChatId))
	if err != nil {
		m.logger.Error("Failed to fetch history from memory layer", err)
	}

	m.logger.Info("Retrieved history from memory layer", history)
	message.History = history

	m.store.SaveItem(partitionKey, sortKey, storedMessage)
	m.logger.Info("Sending message to memory layer", message)

	m.next.SendMessage(message)
}

func (m *MemoryLayer) GetUpdatesChan() chan types.ResponseMessage {
	return m.out
}

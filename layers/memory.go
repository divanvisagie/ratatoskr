package layers

import (
	"ratatoskr/types"
	"time"
)

type MemoryLayer struct {
	// store messages by username and then by timestamp
	store map[string]map[int64]types.StoredMessage
}

func NewMemoryLayer() *MemoryLayer {
	return &MemoryLayer{
		store: make(map[string]map[int64]types.StoredMessage),
	}
}

func (m *MemoryLayer) GetMessages(username string) []types.StoredMessage {
	messages := make([]types.StoredMessage, 0)
	for _, message := range m.store[username] {
		messages = append(messages, message)
	}
	return messages
}

func (m *MemoryLayer) SaveRequestMessage(username string, message string) {
	now := time.Now()
	timestamp := now.Unix()
	if m.store[username] == nil {
		m.store[username] = make(map[int64]types.StoredMessage)
	}
	m.store[username][timestamp] = types.StoredMessage{
		Role:    "user",
		Message: message,
	}
}

func (m *MemoryLayer) SaveResponseMessage(username string, message string) {
	now := time.Now()
	timestamp := now.Unix()
	if m.store[username] == nil {
		m.store[username] = make(map[int64]types.StoredMessage)
	}
	m.store[username][timestamp] = types.StoredMessage{
		Role:    "assistant",
		Message: message,
	}
}

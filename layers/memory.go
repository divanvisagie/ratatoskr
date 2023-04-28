package layers

import (
	"time"
)

type Message struct {
	Role    string
	Message string
}

type MemoryLayer struct {
	// store messages by username and then by timestamp
	store map[string]map[int64]Message
}

func NewMemoryLayer() *MemoryLayer {
	return &MemoryLayer{
		store: make(map[string]map[int64]Message),
	}
}

func (m *MemoryLayer) GetMessages(username string) []Message {
	messages := make([]Message, 0)
	for _, message := range m.store[username] {
		messages = append(messages, message)
	}
	return messages
}

func (m *MemoryLayer) SaveRequestMessage(username string, message string) {
	now := time.Now()
	timestamp := now.Unix()
	if m.store[username] == nil {
		m.store[username] = make(map[int64]Message)
	}
	m.store[username][timestamp] = Message{
		Role:    "user",
		Message: message,
	}
}

func (m *MemoryLayer) SaveResponseMessage(username string, message string) {
	now := time.Now()
	timestamp := now.Unix()
	if m.store[username] == nil {
		m.store[username] = make(map[int64]Message)
	}
	m.store[username][timestamp] = Message{
		Role:    "assistant",
		Message: message,
	}
}

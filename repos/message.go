package repos

import (
	"fmt"
	"ratatoskr/types"
	"sort"
	"time"
)

type Role string

const (
	System    Role = "system"
	User      Role = "user"
	Assistant Role = "assistant"
)

type Message struct {
	store map[string][]types.StoredMessage
}

func NewMessageRepository() *Message {
	return &Message{
		store: make(map[string][]types.StoredMessage),
	}
}

func (m *Message) GetMessages(username string) []types.StoredMessage {
	// messages := make([]types.StoredMessage, 0)
	messages := m.store[username]
	sort.Slice(messages, func(i, j int) bool {
		return messages[i].Timestamp < messages[j].Timestamp
	})

	// only return the last 20 messages
	if len(messages) > 20 {
		messages = messages[len(messages)-20:]
		//some memory management
		m.store[username] = messages[:]
	}

	return messages
}

func (m *Message) SaveMessage(role Role, username string, message string) {

	now := time.Now()
	timestamp := now.UnixMilli()
	if m.store[username] == nil {
		m.store[username] = make([]types.StoredMessage, 0)
	}
	m.store[username] = append(m.store[username], types.StoredMessage{
		Role:      string(role),
		Message:   message,
		Timestamp: timestamp,
	})
	fmt.Printf("Saved message %s from %s at %d\n", message, username, timestamp)
}

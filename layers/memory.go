package layers

import (
	"ratatoskr/types"
	"sync"
	"time"
)

type MemoryLayer struct {
	mu sync.Mutex
	// store messages by username and then by timestamp
	store map[string]map[int64]types.StoredMessage
	child Layer
}

func NewMemoryLayer(child Layer) *MemoryLayer {
	return &MemoryLayer{
		store: make(map[string]map[int64]types.StoredMessage),
		child: child,
	}
}

func (m *MemoryLayer) PassThrough(req *types.RequestMessage) (types.ResponseMessage, error) {
	history := m.getMessages(req.UserName)
	req.Context = history

	res, err := m.child.PassThrough(req)
	if err != nil {
		return types.ResponseMessage{}, err
	}

	m.saveRequestMessage(req.UserName, req.Message)
	m.saveResponseMessage(req.UserName, res.Message)

	return res, nil
}

func (m *MemoryLayer) getMessages(username string) []types.StoredMessage {
	messages := make([]types.StoredMessage, 0)
	for _, message := range m.store[username] {
		messages = append(messages, message)
	}
	return messages[len(messages)-20:] //only return the last 20 messages
}

func (m *MemoryLayer) saveRequestMessage(username string, message string) {
	// limit mutations to one at a time
	// m.mu.Lock()
	// defer m.mu.Unlock()

	now := time.Now()
	timestamp := now.UnixMilli()
	if m.store[username] == nil {
		m.store[username] = make(map[int64]types.StoredMessage)
	}
	m.store[username][timestamp] = types.StoredMessage{
		Role:    "user",
		Message: message,
	}
}

func (m *MemoryLayer) saveResponseMessage(username string, message string) {
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

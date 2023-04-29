package layers

import (
	"ratatoskr/types"
	"sort"
	"time"
)

type MemoryLayer struct {
	// store messages by username and then by timestamp
	store map[string][]types.StoredMessage
	child Layer
}

func NewMemoryLayer(child Layer) *MemoryLayer {
	return &MemoryLayer{
		store: make(map[string][]types.StoredMessage),
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

func (m *MemoryLayer) saveRequestMessage(username string, message string) {

	now := time.Now()
	timestamp := now.UnixMilli()
	if m.store[username] == nil {
		m.store[username] = make([]types.StoredMessage, 0)
	}
	m.store[username] = append(m.store[username], types.StoredMessage{
		Role:      "user",
		Message:   message,
		Timestamp: timestamp,
	})
}

func (m *MemoryLayer) saveResponseMessage(username string, message string) {
	now := time.Now()
	timestamp := now.Unix()
	if m.store[username] == nil {
		m.store[username] = make([]types.StoredMessage, 0)
	}

	m.store[username] = append(m.store[username], types.StoredMessage{
		Role:      "assistant",
		Message:   message,
		Timestamp: timestamp,
	})
}

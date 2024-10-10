package layers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/pkg/types"
	_ "github.com/mattn/go-sqlite3" // Import the SQLite driver
)

type MemoryLayer struct {
	out    chan types.ResponseMessage
	next   types.Cortex
	logger *logger.Logger
	db     *sql.DB
}

func NewMemoryLayer(nextLayer types.Cortex) *MemoryLayer {
	db, err := sql.Open("sqlite3", "./memories.db")
	if err != nil {
		fmt.Println("Failed to open SQLite database:", err)
		return nil
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS documents (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		partitionKey TEXT,
		sortKey TEXT,
		attributes TEXT
	)`)
	if err != nil {
		fmt.Println("Failed to create table:", err)
		return nil
	}

	instance := &MemoryLayer{
		out:    make(chan types.ResponseMessage),
		next:   nextLayer,
		logger: logger.NewLogger("MemoryLayer"),
		db:     db,
	}
	go types.ListenAndRespond(nextLayer, instance.out)
	return instance
}

func (m *MemoryLayer) SendMessage(message types.RequestMessage) {
	now := time.Now().Unix()
	storedMessage := types.StoredMessage{
		Content:   message.Message,
		CreatedAt: now,
	}

	attributesJSON, err := json.Marshal(storedMessage)
	if err != nil {
		m.logger.Error("Failed to marshal JSON:", err)
		return
	}

	partitionKey := fmt.Sprintf("user#%d", message.UserId)
	sortKey := fmt.Sprintf("message#%d#%d", message.ChatId, now)
	_, err = m.db.Exec("INSERT INTO documents (partitionKey, sortKey, attributes) VALUES (?, ?, ?)", partitionKey, sortKey, string(attributesJSON))
	if err != nil {
		m.logger.Error("Failed to insert into SQLite:", err)
		return
	}

	m.logger.Info("Sending message to memory layer", message)

	m.next.SendMessage(message)
}

func (m *MemoryLayer) GetUpdatesChan() chan types.ResponseMessage {
	return m.out
}

func (m *MemoryLayer) GetAttributes(key string) []string {
	var attributesJSON string
	err := m.db.QueryRow("SELECT attributes FROM documents WHERE partitionKey = ?", key).Scan(&attributesJSON)
	if err != nil {
		m.logger.Error("Failed to fetch from SQLite:", err)
		return nil
	}

	var choices []string
	err = json.Unmarshal([]byte(attributesJSON), &choices)
	if err != nil {
		m.logger.Error("Failed to unmarshal JSON:", err)
		return nil
	}

	return choices
}

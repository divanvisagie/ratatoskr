package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/pkg/types"
	_ "modernc.org/sqlite" // Import the SQLite driver
)

const (
	Admin    types.Role = "admin"
	Standard types.Role = "standard"
	Owner    types.Role = "owner"
)

type User struct {
	Id             int64
	TelegramUserId int64
	Username       string
	Role           types.Role
}

/*
Custom Client for a Dynamo like Document Store
using sqlite as a backing storage mechanism
*/
type DocumentStore struct {
	db     *sql.DB
	logger *logger.Logger
}

func NewDocumentStore() *DocumentStore {
	logger := logger.NewLogger("DocumentStore")
	db, err := sql.Open("sqlite", "./memories.db")
	if err != nil {
		logger.Error("Failed to open SQLite database:", err)
		return nil
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS documents (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        partitionKey TEXT,
        sortKey TEXT,
        attributes TEXT
    )`)
	if err != nil {
		logger.Error("Failed to create table:", err)
		return nil
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		telegramUserId INTEGER,
		username TEXT,
		role TEXT
	)`)
	if err != nil {
		logger.Error("Failed to create users table:", err)
		return nil
	}

	// create messages table
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		chatId INTEGER,
		timestamp INTEGER,
		attributes TEXT
	)`)
	if err != nil {
		logger.Error("Failed to create messages table:", err)
		return nil
	}

	return &DocumentStore{
		db:     db,
		logger: logger,
	}
}

func (d *DocumentStore) GetUserByTelegramId(telegramUserId int64) (*User, error) {
	row := d.db.QueryRow("SELECT id, telegramUserId, username, role FROM users WHERE telegramUserId = ?", telegramUserId)
	var user User
	err := row.Scan(&user.Id, &user.TelegramUserId, &user.Username, &user.Role)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		d.logger.Error("Failed to fetch user from SQLite:", err)
		return nil, err
	}
	return &user, nil
}

func (d *DocumentStore) GetUserByTelegramUsername(username string) (*User, error) {
	row := d.db.QueryRow("SELECT id, telegramUserId, username, role FROM users WHERE username = ?", username)
	var user User
	var telegramUserId sql.NullInt64
	err := row.Scan(&user.Id, &telegramUserId, &user.Username, &user.Role)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		d.logger.Error("Failed to fetch user from SQLite:", err)
		return nil, err
	}

	if telegramUserId.Valid {
		user.TelegramUserId = telegramUserId.Int64
	} else {
		user.TelegramUserId = 0 // or any default value you prefer
	}

	return &user, nil
}

func (d *DocumentStore) SaveUser(user User) error {
	if user.Id == 0 {
		_, err := d.db.Exec("INSERT INTO users (telegramUserId, username, role) VALUES (?, ?, ?)", user.TelegramUserId, user.Username, user.Role)
		if err != nil {
			d.logger.Error("Failed to insert into SQLite:", err)
			return err
		}
	} else {
		_, err := d.db.Exec("UPDATE users SET telegramUserId = ?, username = ?, role = ? WHERE id = ?", user.TelegramUserId, user.Username, user.Role, user.Id)
		if err != nil {
			d.logger.Error("Failed to update SQLite:", err)
			return err
		}
	}
	return nil
}

func (d *DocumentStore) GetStoredMessages(chatId int64) ([]types.StoredMessage, error) {
	LIMIT := 15
	rows, err := d.getMessages(chatId, LIMIT)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []types.StoredMessage
	for rows.Next() {
		var attributesJSON string
		var timestamp int64
		if err := rows.Scan(&timestamp, &attributesJSON); err != nil {
			d.logger.Error("Failed to scan row:", err)
			return nil, err
		}

		var result types.StoredMessage
		if err := json.Unmarshal([]byte(attributesJSON), &result); err != nil {
			d.logger.Error("Failed to unmarshal JSON:", err)
			return nil, err
		}

		results = append(results, result)
	}

	if err := rows.Err(); err != nil {
		d.logger.Error("Rows iteration error:", err)
		return nil, err
	}

	for i, j := 0, len(results)-1; i < j; i, j = i+1, j-1 {
		results[i], results[j] = results[j], results[i]
	}

	return results, nil
}

func (d *DocumentStore) getMessages(chatId int64, limit int) (*sql.Rows, error) {
	query := `
		SELECT timestamp, attributes FROM messages 
		WHERE chatId = ?
		ORDER BY timestamp DESC LIMIT ?`

	rows, err := d.db.Query(query, chatId, limit)
	if err != nil {
		d.logger.Error("Failed to fetch from SQLite:", err)
		return nil, err
	}

	return rows, nil
}

func (s *DocumentStore) FetchMessagesByIDs(ids []int64) ([]types.StoredMessage, error) {
	if len(ids) == 0 {
		return nil, fmt.Errorf("no IDs provided")
	}

	// Dynamically build the IN clause with the right number of placeholders
	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = "?"      // SQLite placeholder for arguments
		args[i] = id               // Assign each ID to args slice
	}

	// Join the placeholders into the query
	query := fmt.Sprintf("SELECT id, content FROM messages WHERE id IN (%s)", strings.Join(placeholders, ","))

	// Execute the query with the dynamically generated placeholders
	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}
	defer rows.Close()

	var messages []types.StoredMessage
	for rows.Next() {
		var message types.StoredMessage
		if err := rows.Scan(&message.Content); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// Save the message and return the id with which it was saved
func (d *DocumentStore) SaveMessage(chatId int64, timestamp int64, item types.StoredMessage) (int64, error) {
	attributesJSON, err := json.Marshal(item)
	if err != nil {
		d.logger.Error("Failed to marshal JSON:", err)
		return 0, err
	}

	result, err := d.db.Exec("INSERT INTO messages (chatId, timestamp, attributes) VALUES (?, ?, ?)", chatId, timestamp, string(attributesJSON))
	if err != nil {
		d.logger.Error("Failed to insert into SQLite:", err)
		return 0, err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		d.logger.Error("Failed to retrieve inserted id:", err)
		return 0, err
	}

	return id, nil
}

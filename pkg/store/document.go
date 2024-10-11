package store

import (
	"database/sql"
	"encoding/json"

	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/pkg/types"
	_ "modernc.org/sqlite" // Import the SQLite driver
)

type Role string

const (
	Admin    Role = "admin"
	Standard Role = "standard"
	Owner    Role = "owner"
)

type User struct {
	TelegramUserId int
	Username       string
	Role           Role
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

	return &DocumentStore{
		db:     db,
		logger: logger,
	}
}

func (d *DocumentStore) GetUserByTelegramId(telegramUserId int64) (*User, error) {
	row := d.db.QueryRow("SELECT telegramUserId, username, role FROM users WHERE telegramUserId = ?", telegramUserId)
	var user User
	err := row.Scan(&user.TelegramUserId, &user.Username, &user.Role)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		d.logger.Error("Failed to fetch user from SQLite:", err)
		return nil, err
	}
	return &user, nil
}

func (d *DocumentStore) GetUserByTelegramUsername(username string) (*User, error) {
	row := d.db.QueryRow("SELECT telegramUserId, username, role FROM users WHERE username = ?", username)
	var user User
	var telegramUserId sql.NullInt64
	err := row.Scan(&telegramUserId, &user.Username, &user.Role)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		d.logger.Error("Failed to fetch user from SQLite:", err)
		return nil, err
	}

	if telegramUserId.Valid {
		user.TelegramUserId = int(telegramUserId.Int64)
	} else {
		user.TelegramUserId = 0 // or any default value you prefer
	}

	return &user, nil
}

func (d *DocumentStore) GetStoredMessages(partitionKey string, sortKey string) ([]types.StoredMessage, error) {
	rows, err := d.fetchItems(partitionKey, sortKey)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []types.StoredMessage
	for rows.Next() {
		var attributesJSON string
		if err := rows.Scan(&attributesJSON); err != nil {
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

	return results, nil
}

func (d *DocumentStore) fetchItems(partitionKey string, sortKey string) (*sql.Rows, error) {
	rows, err := d.db.Query("SELECT attributes FROM documents WHERE partitionKey = ? AND sortKey LIKE ?", partitionKey, sortKey+"%")
	if err != nil {
		d.logger.Error("Failed to fetch from SQLite:", err)
		return nil, err
	}
	return rows, nil
}

func (d *DocumentStore) SaveItem(partitionKey string, sortKey string, item interface{}) {
	attributesJSON, err := json.Marshal(item)
	if err != nil {
		d.logger.Error("Failed to marshal JSON:", err)
		return
	}

	_, err = d.db.Exec("INSERT INTO documents (partitionKey, sortKey, attributes) VALUES (?, ?, ?)", partitionKey, sortKey, string(attributesJSON))
	if err != nil {
		d.logger.Error("Failed to insert into SQLite:", err)
		return
	}
}

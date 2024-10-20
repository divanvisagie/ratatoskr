package layers

import (
	"strings"
	"time"

	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/internal/services"
	"github.com/divanvisagie/ratatoskr/pkg/clients"
	"github.com/divanvisagie/ratatoskr/pkg/store"
	"github.com/divanvisagie/ratatoskr/pkg/types"
)

type MemoryLayer struct {
	out    chan types.ResponseMessage
	next   types.Cortex
	logger *logger.Logger
	store  *store.DocumentStore
	cfg    config.Config
}

func (m *MemoryLayer) listenAndRespond() {
	for response := range m.next.GetUpdatesChan() {
		storedMessage := types.StoredMessage{
			Content:   response.Message,
			Role:      "assistant",
			CreatedAt: time.Now().Unix(),
		}
		m.store.SaveMessage(response.ChatId, storedMessage.CreatedAt, storedMessage)

		m.out <- response
	}
}

func NewMemoryLayer(nextLayer types.Cortex, cfg config.Config) *MemoryLayer {
	instance := &MemoryLayer{
		out:    make(chan types.ResponseMessage),
		next:   nextLayer,
		logger: logger.NewLogger("MemoryLayer"),
		store:  store.NewDocumentStore(),
		cfg:    cfg,
	}
	go instance.listenAndRespond()
	return instance
}

func (m *MemoryLayer) Tell(message types.RequestMessage) {
	now := time.Now().Unix()
	storedMessage := types.StoredMessage{
		Content:   message.Message,
		Role:      "user",
		CreatedAt: now,
		Username:  message.Username,
		Fullname:  message.Fullname,
	}

	history, err := m.store.GetStoredMessages(message.ChatId)
	if err != nil {
		m.logger.Error("Failed to fetch history from memory layer", err)
	}

	message.History = history

	m.store.SaveMessage(message.ChatId, now, storedMessage)
	ec := clients.NewEmbeddingsClient(m.cfg.OpenAIKey)
	ltms := services.NewLongTermMemoryService(ec, m.cfg)
	// TODO: Make sure we actually store with an Id
	ltms.StoreMessageLongTerm(0, storedMessage)

	// Ignore group chat messages unless the bot's username is mentioned
	if message.AuthUser.TelegramUserId < 0 && !strings.Contains(message.Message, m.cfg.BotUsername) {
		m.logger.Info("Ignoring group chat message")
		return
	}
	m.next.Tell(message)
}

func (m *MemoryLayer) GetUpdatesChan() chan types.ResponseMessage {
	return m.out
}

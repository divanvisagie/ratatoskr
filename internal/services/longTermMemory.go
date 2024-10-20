package services

import (
	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/pkg/clients"
	"github.com/divanvisagie/ratatoskr/pkg/store"
	"github.com/divanvisagie/ratatoskr/pkg/types"
)

type LongTermMemoryService struct {
	store            store.DocumentStore
	embeddingsClient clients.EmbeddingsClient
	logger           logger.Logger
	chromaURL        string
	cfg              config.Config
}

func NewLongTermMemoryService(client clients.EmbeddingsClient, cfg config.Config) *LongTermMemoryService {
	return &LongTermMemoryService{
		store:            *store.NewDocumentStore(),
		embeddingsClient: client,
		logger:           *logger.NewLogger("LongTermMemoryService"),
		chromaURL:        cfg.ChromaBaseUrl,
		cfg:              cfg,
	}
}

// StoreMessageLongTerm generates embeddings and stores them in Chroma
func (l *LongTermMemoryService) StoreMessageLongTerm(id int64, chatId int64, message types.StoredMessage) {

	cc := clients.NewChromaClient(l.cfg, chatId)
	err := cc.SaveEmbeddedVector(id, chatId, message.Content)
	if err != nil {
		// Handle error
		l.logger.Error("Failed to save embedding in Chroma", err)
		return
	}

}

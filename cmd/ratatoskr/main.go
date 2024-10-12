package main

import (
	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/pkg/store"
	"github.com/divanvisagie/ratatoskr/pkg/telegram"
)

func initialSetup(cfg *config.Config) {
	s := store.NewDocumentStore()	

	usr := store.User{
		Username: cfg.Owner,
	}

	s.SaveUser(usr)
}

func main() {
	cfg := config.LoadConfig()

	initialSetup(cfg)

	telegram.StartBot(cfg.TelegramToken, cfg)
}

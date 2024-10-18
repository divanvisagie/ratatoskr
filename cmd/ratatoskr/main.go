package main

import (
	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/pkg/store"
	"github.com/divanvisagie/ratatoskr/pkg/telegram"
)

func initialSetup(cfg *config.Config) {
	l := logger.NewLogger("main")
	s := store.NewDocumentStore()

	user, err := s.GetUserByTelegramUsername(cfg.Owner)
	if err != nil {
		l.Error("Error getting user", err)
		panic(err)
	}

	if user == nil {
		l.Info("User not found, creating")

		usr := store.User{
			Username: cfg.Owner,
			Role:     "owner",
		}

		l.Info("Saving user", usr)

		err := s.SaveUser(usr)
		if err != nil {
			l.Error("Error saving user", err)
			panic(err)
		}
	} else {
		l.Info("User found", user)
	}
}

func main() {
	cfg := config.LoadConfig()

	initialSetup(cfg)

	telegram.StartBot(cfg.TelegramToken, cfg)
}

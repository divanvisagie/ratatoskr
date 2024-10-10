package main

import (
	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/pkg/telegram"
)

func main() {
	cfg := config.LoadConfig()

	telegram.StartBot(cfg.TelegramToken, cfg)
}

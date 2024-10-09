package main

import (
	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/pkg/openai"
	"github.com/divanvisagie/ratatoskr/pkg/telegram"
)

func main() {
	cfg := config.LoadConfig()

	openaiClient := openai.NewClient(cfg.OpenAIKey)

	telegram.StartBot(cfg.TelegramToken, openaiClient)
}

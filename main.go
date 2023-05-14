package main

import (
	"log"
	"os"
	"ratatoskr/handler"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)
func main() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panicf("Failed to start: %v with token", err)
	}
	// bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)

	updates := bot.GetUpdatesChan(u)

	handler := handler.NewHandler(bot)

	for update := range updates {
		if update.Message != nil { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			handler.HandleTelegramMessages(update)
		}
	}
}

package telegram

import (
	"github.com/divanvisagie/ratatoskr/pkg/openai" 
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
)

func StartBot(token string, openaiClient *openai.Client) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			prompt := update.Message.Text
			completion, err := openaiClient.GetCompletion(prompt)
			if err != nil {
				log.Println("Error fetching OpenAI completion:", err)
				continue
			}

			typing := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
			bot.Send(typing)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, completion)
			bot.Send(msg)
		}
	}
}

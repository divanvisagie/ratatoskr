package telegram

import (
	"log"

	"github.com/divanvisagie/ratatoskr/internal/config/layers"
	"github.com/divanvisagie/ratatoskr/pkg/openai"
	"github.com/divanvisagie/ratatoskr/pkg/types"
	"github.com/go-telegram-bot-api/telegram-bot-api/v5"
)
func listenAndRespond(bot *tgbotapi.BotAPI, securityLayer layers.SecurityLayer) {
	for {
		select {
		case response := <-securityLayer.GetUpdatesChan():
            log.Printf("Sending message to chat %d: %s\n", response.ChatId, response.Content)
			msg := tgbotapi.NewMessage(response.ChatId, response.Content)
			bot.Send(msg)
		}
	}
}
func StartBot(token string, openaiClient *openai.Client) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	firstLayer := layers.NewSecurityLayer()

    go listenAndRespond(bot, *firstLayer)

	for update := range updates {
		if update.Message != nil {

			requestMessage := types.RequestMessage{
                ChatId: update.Message.Chat.ID,
				Content: update.Message.Text,
			}

			firstLayer.SendMessage(requestMessage)

			typing := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
			bot.Send(typing)
		}
	}
}

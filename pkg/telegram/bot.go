package telegram

import (
	"log"

	"github.com/divanvisagie/ratatoskr/internal/capabilities"
	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/internal/layers"
	"github.com/divanvisagie/ratatoskr/pkg/types"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func listenAndRespond(bot *tgbotapi.BotAPI, firstLayer types.Cortex) {
	for response := range firstLayer.GetUpdatesChan() {
		log.Printf("Sending message to chat %d: %s\n", response.ChatId, response.Message)
		msg := tgbotapi.NewMessage(response.ChatId, response.Message)
		bot.Send(msg)
	}
}

func StartBot(token string, cfg *config.Config) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	chat := capabilities.NewChatCapability(cfg)
	image := capabilities.NewImageGenerationCapability(cfg)
	caps := []types.Capability{chat, image}

	selectionLayer := layers.NewSelectionLayer(*cfg, &caps)
	memoryLayer := layers.NewMemoryLayer(selectionLayer)
	securityLayer := layers.NewSecurityLayer(memoryLayer)

	go listenAndRespond(bot, securityLayer)

	// Listen for messages on the input channel
	for update := range updates {
		if update.Message != nil {
			requestMessage := types.RequestMessage{
				ChatId:  update.Message.Chat.ID,
				Message: update.Message.Text,
			}

			go securityLayer.SendMessage(requestMessage)

			typing := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
			bot.Send(typing)
		}
	}
}

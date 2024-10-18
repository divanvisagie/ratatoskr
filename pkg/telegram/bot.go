package telegram

import (
	"log"

	"github.com/divanvisagie/ratatoskr/internal/capabilities"
	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/internal/layers"
	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/pkg/types"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func listenAndRespond(bot *tgbotapi.BotAPI, firstLayer types.Cortex, logger *logger.Logger) {
	for response := range firstLayer.GetUpdatesChan() {
		if response.Data != nil {
			file := tgbotapi.FileBytes{
				Name:  "image.jpg",
				Bytes: response.Data,
			}
			photo := tgbotapi.NewPhoto(response.ChatId, file)
			bot.Send(photo)
		} else {
			logger.Info("Sending message to chat", response)
			msg := tgbotapi.NewMessage(response.ChatId, response.Message)
			bot.Send(msg)
		}
	}
}

func StartBot(token string, cfg *config.Config) {
	contextLogger := logger.NewLogger("TelegramBot")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	test := capabilities.NewTestCapability()
	chat := capabilities.NewChatCapability(cfg)
	image := capabilities.NewImageGenerationCapability(cfg)
	caps := []types.Capability{chat, image, test}

	selectionLayer := layers.NewSelectionLayer(*cfg, &caps)
	memoryLayer := layers.NewMemoryLayer(selectionLayer)
	securityLayer := layers.NewSecurityLayer(memoryLayer)

	go listenAndRespond(bot, securityLayer, contextLogger)

	// Listen for messages on the input channel
	for update := range updates {
		if update.Message != nil {
			au := types.AuthUser{}

			if update.Message.Chat.IsGroup() {
				au.TelegramUserId = update.Message.Chat.ID
				au.ChatName = update.Message.Chat.Title
			} else {
				au.TelegramUserId = update.Message.From.ID
				au.ChatName = update.Message.From.UserName
			}

			requestMessage := types.RequestMessage{
				UserId:   update.Message.From.ID,
				ChatId:   update.Message.Chat.ID,
				Message:  update.Message.Text,
				AuthUser: au,
			}

			go securityLayer.SendMessage(requestMessage)

			typing := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
			bot.Send(typing)
		}
	}
}

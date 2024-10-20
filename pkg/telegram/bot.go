package telegram

import (
	"fmt"
	"log"

	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/internal/layers"
	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/pkg/types"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func listenAndRespond(bot *tgbotapi.BotAPI, firstLayer types.Cortex, logger *logger.Logger) {
	for res := range firstLayer.GetUpdatesChan() {
		switch res.DataType {
		case types.JPG:
			file := tgbotapi.FileBytes{
				Name:  "image.jpg",
				Bytes: res.Data,
			}
			photo := tgbotapi.NewPhoto(res.ChatId, file)
			bot.Send(photo)

		case types.MP3:
			// code to handle MP3 data type goes here
			file := tgbotapi.FileBytes{
				Name:  "image.mp3",
				Bytes: res.Data,
			}
			mp3 := tgbotapi.NewVoice(res.ChatId, file)
			bot.Send(mp3)

		case types.BUSY:
			typing := tgbotapi.NewChatAction(res.ChatId, tgbotapi.ChatTyping)
			bot.Send(typing)

		default:
			logger.Info("Sending message to chat")
			msg := tgbotapi.NewMessage(res.ChatId, res.Message)
			msg.ParseMode = "Markdown"
			bot.Send(msg)
		}
	}
}

func listenToBusy(busyChannel chan types.BusyIndicatorMessage, bot *tgbotapi.BotAPI, logger *logger.Logger) {
	for b := range busyChannel {
		logger.Info("Bot is busy")
		typing := tgbotapi.NewChatAction(b.ChatId, tgbotapi.ChatTyping)
		bot.Send(typing)
	}
}

func StartBot(token string, cfg *config.Config) {
	logger := logger.NewLogger("TelegramBot")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	busyChannel := make(chan types.BusyIndicatorMessage)

	selectionLayer := layers.NewSelectionLayer(*cfg, busyChannel)
	memoryLayer := layers.NewMemoryLayer(selectionLayer, *cfg)
	securityLayer := layers.NewSecurityLayer(memoryLayer, *cfg)

	go listenAndRespond(bot, securityLayer, logger)
	go listenToBusy(busyChannel, bot, logger)

	// Listen for messages on the input channel
	for update := range updates {
		if update.Message != nil {
			au := types.AuthUser{}

			if update.Message.Chat.ID < 0 {
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
				Username: update.Message.From.UserName,
				AuthUser: au,
				Fullname: fmt.Sprintf("%s %s", update.Message.From.FirstName, update.Message.From.LastName),
			}

			go securityLayer.Tell(requestMessage)
		}
	}
}

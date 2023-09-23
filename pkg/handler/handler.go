package handler

import (
	"fmt"
	"log"
	caps "ratatoskr/pkg/capabilities"
	"ratatoskr/pkg/layers"
	"ratatoskr/pkg/repos"
	"ratatoskr/pkg/types"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot          *tgbotapi.BotAPI
	gatewayLayer layers.Layer
}

func NewHandler(bot *tgbotapi.BotAPI) *Handler {
	memRepo := repos.NewMessageRepository()

	caps := []types.Capability{
		caps.NewTestRedis(memRepo),
		caps.NewMemoryDump(memRepo),
		caps.NewMemoryWipe(memRepo),
		caps.NewNotion(),
		caps.NewLinkProcessor(memRepo),
		caps.NewChatGPT(),
	}

	//build up the layers
	capabilityLayer := layers.NewCapabilitySelector(caps)
	memoryLayer := layers.NewMemoryLayer(memRepo, capabilityLayer)
	securityLayer := layers.NewSecurity(memoryLayer)

	return &Handler{bot: bot, gatewayLayer: securityLayer}
}

func (h *Handler) HandleTelegramMessages(update tgbotapi.Update) {
	if update.Message != nil {
		if update.Message.Text == "/menu" {
			sendMenu(h.bot, update.Message.Chat.ID)
			return
		}

		//simulate typing
		typingMsg := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
		h.bot.Send(typingMsg)

		//pass through the gateway layer
		req := &types.RequestMessage{
			ChatID:   update.Message.Chat.ID,
			Message:  update.Message.Text,
			UserName: update.Message.From.UserName,
		}
		res, err := h.gatewayLayer.PassThrough(req)
		if err != nil {
			log.Println(err)
			msg := tgbotapi.NewMessage(res.ChatID, "Error while processing message")
			msg.ReplyToMessageID = update.Message.MessageID

			h.bot.Send(msg)
		}

		if res.Bytes != nil {
			now := time.Now()
			timestamp := now.UnixMilli()
			key := fmt.Sprintf("%d", timestamp)
			csvFile := tgbotapi.FileBytes{Name: key + ".csv", Bytes: res.Bytes}
			msg := tgbotapi.NewDocument(res.ChatID, csvFile)
			h.bot.Send(msg)
			//TODO
		} else {
			// Respond to the user
			msg := tgbotapi.NewMessage(res.ChatID, res.Message)
			msg.ParseMode = "markdown"
			h.bot.Send(msg)
		}
	}
}

// Function to send custom keyboard with menu options
func sendMenu(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, "Select an option:")
	msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Clear memory"),
			tgbotapi.NewKeyboardButton("Memory dump"),
		),
	)
	bot.Send(msg)
}

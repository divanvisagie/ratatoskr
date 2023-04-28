package handler

import (
	"ratatoskr/capabilities"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot          *tgbotapi.BotAPI
	capabilities []capabilities.Capability
}

func NewHandler(bot *tgbotapi.BotAPI) *Handler {
	capabilities := []capabilities.Capability{
		capabilities.ChatGPT{},
	}
	return &Handler{bot, capabilities}
}

// function to handle telegram messages
func (h *Handler) HandleTelegramMessages(update tgbotapi.Update) {
	if update.Message != nil { // If we got a message

		//simulate typing
		typingMsg := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
		h.bot.Send(typingMsg)

		//check capabilities
		for _, capability := range h.capabilities {
			if capability.Check(update) {
				response := capability.Execute(update)
				msg := tgbotapi.NewMessage(response.ChatID, response.Message)
				msg.ReplyToMessageID = update.Message.MessageID
				h.bot.Send(msg)
			}
		}
	}
}

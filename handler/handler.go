package handler

import (
	"log"
	"ratatoskr/capabilities"
	"ratatoskr/layers"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot          *tgbotapi.BotAPI
	capabilities []capabilities.Capability
	memoryLayer  *layers.MemoryLayer
}

func NewHandler(bot *tgbotapi.BotAPI) *Handler {
	memoryLayer := layers.NewMemoryLayer()
	capabilities := []capabilities.Capability{
		capabilities.NewChatGPT(memoryLayer),
	}
	return &Handler{bot, capabilities, memoryLayer}
}

func (h *Handler) HandleTelegramMessages(update tgbotapi.Update) {
	if update.Message != nil {
		//simulate typing
		typingMsg := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
		h.bot.Send(typingMsg)

		//check capabilities
		for _, capability := range h.capabilities {
			if capability.Check(update) {
				response, err := capability.Execute(update)
				if err != nil {
					log.Println(err)
					msg := tgbotapi.NewMessage(response.ChatID, "Error while processing message")
					msg.ReplyToMessageID = update.Message.MessageID

					h.bot.Send(msg)
					continue
				}

				h.memoryLayer.SaveRequestMessage(
					update.Message.From.UserName,
					update.Message.Text,
				)
				h.memoryLayer.SaveResponseMessage(
					update.Message.From.UserName,
					response.Message,
				)

				msg := tgbotapi.NewMessage(response.ChatID, response.Message)
				msg.ReplyToMessageID = update.Message.MessageID
				msg.ParseMode = "markdown"

				h.bot.Send(msg)
			}
		}
	}
}

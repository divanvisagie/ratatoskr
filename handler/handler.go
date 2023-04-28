package handler

import (
	"log"
	"ratatoskr/capabilities"
	"ratatoskr/layers"
	"ratatoskr/types"

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

func executeCorrectCapability(update tgbotapi.Update, capabilities []capabilities.Capability) (res types.ResponseMessage, err error) {
	for _, capability := range capabilities {
		if capability.Check(update) {
			return capability.Execute(update)
		}
	}
	return types.ResponseMessage{}, nil
}

func (h *Handler) HandleTelegramMessages(update tgbotapi.Update) {
	if update.Message != nil {
		//simulate typing
		typingMsg := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
		h.bot.Send(typingMsg)

		res, err := executeCorrectCapability(update, h.capabilities)
		if err != nil {
			log.Println(err)
			msg := tgbotapi.NewMessage(res.ChatID, "Error while processing message")
			msg.ReplyToMessageID = update.Message.MessageID

			h.bot.Send(msg)
		}

		h.memoryLayer.SaveRequestMessage(
			update.Message.From.UserName,
			update.Message.Text,
		)
		h.memoryLayer.SaveResponseMessage(
			update.Message.From.UserName,
			res.Message,
		)

		msg := tgbotapi.NewMessage(res.ChatID, res.Message)
		msg.ReplyToMessageID = update.Message.MessageID
		msg.ParseMode = "markdown"

		h.bot.Send(msg)
	}
}

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
	gatewayLayer layers.Layer
}

func NewHandler(bot *tgbotapi.BotAPI) *Handler {
	capabilities := []types.Capability{
		capabilities.NewNotion(),
		capabilities.NewLinkProcessor(),
		capabilities.NewChatGPT(),
	}

	//build up the layers
	capabilityLayer := layers.NewCapabilitySelector(capabilities)
	memoryLayer := layers.NewMemoryLayer(capabilityLayer)
	securityLayer := layers.NewSecurity(memoryLayer)

	return &Handler{bot: bot, gatewayLayer: securityLayer}
}

func (h *Handler) HandleTelegramMessages(update tgbotapi.Update) {
	if update.Message != nil {
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

		msg := tgbotapi.NewMessage(res.ChatID, res.Message)
		msg.ParseMode = "markdown"

		h.bot.Send(msg)
	}
}

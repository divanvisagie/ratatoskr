package main

import (
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot *tgbotapi.BotAPI
}

// function to handle telegram messages
func (h *Handler) HandleTelegramMessages(update tgbotapi.Update) {
	if update.Message != nil { // If we got a message
		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		//simulate typing
		typingMsg := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
		h.bot.Send(typingMsg)

		// wait
		time.Sleep(5 * time.Second)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		h.bot.Send(msg)
	}
}

package capabilities

import (
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ChatGPT struct{}

func (c ChatGPT) Check(update tgbotapi.Update) bool {
	return true
}

func (c ChatGPT) Execute(update tgbotapi.Update) ResponseMessage {
	// simulate typing
	// typingMsg := tgbotapi.NewChatAction(update.Message.Chat.ID, tgbotapi.ChatTyping)
	// bot.Send(typingMsg)

	// wait
	time.Sleep(5 * time.Second)

	// msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
	// msg.ReplyToMessageID = update.Message.MessageID

	// bot.Send(msg)

	return ResponseMessage{
		ChatID:  update.Message.Chat.ID,
		Message: "Hello, I'm a bot.",
	}
}

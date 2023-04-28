package capabilities

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type ResponseMessage struct {
	ChatID  int64
	Message string
}

type RequestMessage struct {
	ChatID  int64
	Message string
}

type Capability interface {
	Check(update tgbotapi.Update) bool
	Execute(update tgbotapi.Update) (res ResponseMessage, err error)
}

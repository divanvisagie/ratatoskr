package capabilities

import (
	"ratatoskr/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Capability interface {
	Check(update tgbotapi.Update) bool
	Execute(update tgbotapi.Update) (res types.ResponseMessage, err error)
}

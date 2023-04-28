package capabilities

import (
	clients "ratatoskr/clients"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type ChatGPT struct {
	systemPrompt string
}

func NewChatGPT() *ChatGPT {
	return &ChatGPT{
		systemPrompt: `You are Ratatoskr, 
		an EI (Extended Intelligence). 
		An extended intelligence is a software system 
		that utilises multiple Language Models, AI models, 
		NLP Functions and other capabilities to best serve 
		the user.`,
	}
}

func (c ChatGPT) Check(update tgbotapi.Update) bool {
	return true
}

func (c ChatGPT) Execute(update tgbotapi.Update) ResponseMessage {
	cl := clients.NewOpenAIClient(c.systemPrompt).SetMaxTokens(500)
	res := cl.Complete(update.Message.Text)

	return ResponseMessage{
		ChatID:  update.Message.Chat.ID,
		Message: res,
	}
}

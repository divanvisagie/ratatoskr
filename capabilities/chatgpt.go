package capabilities

import (
	clients "ratatoskr/clients"
	layer "ratatoskr/layers"
	"ratatoskr/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	openai "github.com/sashabaranov/go-openai"
)

type ChatGPT struct {
	systemPrompt string
	memoryLayer  *layer.MemoryLayer
}

func NewChatGPT(memoryLayer *layer.MemoryLayer) *ChatGPT {
	return &ChatGPT{
		systemPrompt: `You are Ratatoskr, 
		an EI (Extended Intelligence). 
		An extended intelligence is a software system 
		that utilises multiple Language Models, AI models, 
		NLP Functions and other capabilities to best serve 
		the user.`,
		memoryLayer: memoryLayer,
	}
}

func messageToChatCompletionMessage(message types.StoredMessage) openai.ChatCompletionMessage {
	if message.Role == "user" {
		return openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: message.Message,
		}
	} else {
		return openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: message.Message,
		}
	}
}

func (c ChatGPT) Check(update tgbotapi.Update) bool {
	return true
}

func (c ChatGPT) Execute(update tgbotapi.Update) (res types.ResponseMessage, err error) {

	previousMessages := c.memoryLayer.GetMessages(update.Message.From.UserName)
	history := make([]openai.ChatCompletionMessage, len(previousMessages))
	for i, message := range previousMessages {
		history[i] = messageToChatCompletionMessage(message)
	}

	cl := clients.NewOpenAIClient(c.systemPrompt).
		SetHistory(history).
		SetMaxTokens(500)

	message := cl.Complete(update.Message.Text)

	rm := types.ResponseMessage{
		ChatID:  update.Message.Chat.ID,
		Message: message,
	}
	return rm, nil
}

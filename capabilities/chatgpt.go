package capabilities

import (
	clients "ratatoskr/clients"
	"ratatoskr/types"

	openai "github.com/sashabaranov/go-openai"
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
		the user.
		
		Your code lives at https://github.com/divanvisagie/Ratatoskr
		`,
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

func (c ChatGPT) Check(req *types.RequestMessage) bool {
	return true
}

func (c ChatGPT) Execute(req *types.RequestMessage) (types.ResponseMessage, error) {

	previousMessages := req.Context
	history := make([]openai.ChatCompletionMessage, len(previousMessages))
	for i, message := range previousMessages {
		history[i] = messageToChatCompletionMessage(message)
	}

	client := clients.NewOpenAIClient(c.systemPrompt).
		SetHistory(history).
		SetMaxTokens(500)

	message := client.Complete(req.Message)

	res := types.ResponseMessage{
		ChatID:  req.ChatID,
		Message: message,
	}
	return res, nil
}

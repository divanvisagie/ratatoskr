package caps

import (
	client "ratatoskr/pkg/client"
	"ratatoskr/pkg/types"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type ChatGPT struct {
	systemPrompt string
}

func NewChatGPT() *ChatGPT {
	return &ChatGPT{
		systemPrompt: strings.TrimSpace(`
		Ratatoskr is an EI (Extended Intelligence) written in Go. 
		An extended intelligence is a software system 
		that utilises multiple Language Models, AI models, 
		NLP Functions and other capabilities to best serve 
		the user.

		As the response Model for Ratatoskr, you answer user questions as if you are the main
		brain of the system. 
		
		If a user asks about how you work or your code, 
		respond with the following link: https://github.com/divanvisagie/Ratatoskr
		`),
	}
}

func messageToChatCompletionMessage(message types.StoredMessage) openai.ChatCompletionMessage {
	if message.Role == "user" {
		return openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: message.Message,
		}
	} else if message.Role == "assistant" {
		return openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: message.Message,
		}
	} else {
		return openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleSystem,
			Content: message.Message,
		}
	}
}

func (c ChatGPT) Check(req *types.RequestMessage) float32 {
	return 0.9
}

func (c ChatGPT) Execute(req *types.RequestMessage) (types.ResponseMessage, error) {

	previousMessages := req.Context
	history := make([]openai.ChatCompletionMessage, len(previousMessages))
	for i, message := range previousMessages {
		history[i] = messageToChatCompletionMessage(message)
	}

	client := client.NewOpenAIClient().
		AddSystemMessage(c.systemPrompt).
		SetHistory(history).
		SetMaxTokens(500).
		AddUserMessage(req.Message)

	message := client.Complete()

	res := types.ResponseMessage{
		ChatID:  req.ChatID,
		Message: message,
	}
	return res, nil
}

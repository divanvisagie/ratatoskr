package capabilities

import (
	_ "embed"
	"fmt"

	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/pkg/clients"
	"github.com/divanvisagie/ratatoskr/pkg/types"

	o "github.com/sashabaranov/go-openai"
)

type ChatCapability struct {
	cfg    *config.Config
	out    chan types.ResponseMessage
	done   chan bool
	logger *logger.Logger
}

//go:embed prompt.txt
var systemPrompt string

// create constructor function
func NewChatCapability(cfg *config.Config) *ChatCapability {
	// load prompt string from file prompt.txt and include at build time
	instance := &ChatCapability{
		cfg:    cfg,
		out:    make(chan types.ResponseMessage),
		done:   make(chan bool),
		logger: logger.NewLogger("ChatCapability"),
	}

	go types.ListenAndRespond(instance, instance.out)

	return instance
}

func (c *ChatCapability) Tell(msg types.RequestMessage) {
	c.logger.Info("Sending message to chat capability", msg)
	client := clients.NewChatClient(c.cfg.OpenAIKey)

	if msg.AuthUser.TelegramUserId < 0 {
		systemPrompt = fmt.Sprintf("%s\nFYI: You are in a group chat.", systemPrompt)
	}
	message := fmt.Sprintf("@%s (%s): %s", msg.Username, msg.Fullname, msg.Message)

	client.SetSystemPrompt(systemPrompt)
	client.AddStoredMessages(msg.History)
	client.AddMessage("user", message)
	response, err := client.GetCompletion()

	if err != nil {
		c.out <- types.ResponseMessage{
			UserId:  msg.UserId,
			ChatId:  msg.ChatId,
			Message: "I'm sorry, I'm having trouble processing your request",
		}
	} else {
		c.out <- types.ResponseMessage{
			UserId:  msg.UserId,
			ChatId:  msg.ChatId,
			Message: response,
		}
	}
}

func (c *ChatCapability) ReceiveMessage() types.ResponseMessage {
	return <-c.out
}

func (c *ChatCapability) GetUpdatesChan() chan types.ResponseMessage {
	return c.out
}

func (c *ChatCapability) Stop() {
	c.done <- true
}

func (c *ChatCapability) Describe() o.Tool {
	fd := o.FunctionDefinition{
		Name:        "ChatCapability",
		Description: "General chat capability that sends responses from gpt-4o",
	}
	return o.Tool{
		Type:     o.ToolTypeFunction,
		Function: &fd,
	}
}

func (c *ChatCapability) Check(inputMessage types.RequestMessage) float64 {
	return 0.0
}

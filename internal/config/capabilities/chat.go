package capabilities

import (
	_ "embed"
	"log"

	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/pkg/openai"
	"github.com/divanvisagie/ratatoskr/pkg/types"

	o "github.com/sashabaranov/go-openai"
)

type ChatCapability struct {
	cfg  *config.Config
	out  chan types.ResponseMessage
	done chan bool
}

//go:embed prompt.txt
var systemPrompt string

// create constructor function
func NewChatCapability(cfg *config.Config) *ChatCapability {
	// load prompt string from file prompt.txt and include at build time

	instance := &ChatCapability{
		cfg:  cfg,
		out:  make(chan types.ResponseMessage),
		done: make(chan bool),
	}

	go types.ListenAndRespond(instance, instance.out)

	return instance
}


func (c *ChatCapability) SendMessage(msg types.RequestMessage) {
	log.Println("Chat Capability: ", msg)
	client := openai.NewClient(c.cfg.OpenAIKey)
	client.SetSystemPrompt(systemPrompt)
	response, err := client.GetCompletion(msg.Message)

	if err != nil {
		c.out <- types.ResponseMessage{
			ChatId:  msg.ChatId,
			Message: "I'm sorry, I'm having trouble processing your request",
		}
	} else {
		c.out <- types.ResponseMessage{
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

func (c *ChatCapability) Describe() o.Tool{
	fd := o.FunctionDefinition{
		Name: "ChatCapability",
	    Description: "General chat capability that sends responses from gpt-4o",
	}
	return o.Tool{
		Type: o.ToolTypeFunction,
		Function: &fd,
	}
}

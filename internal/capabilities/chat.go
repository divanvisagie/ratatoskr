package capabilities

import (
	_ "embed"
	"fmt"

	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/pkg/clients"
	"github.com/divanvisagie/ratatoskr/pkg/store"
	"github.com/divanvisagie/ratatoskr/pkg/types"

	"github.com/sashabaranov/go-openai"
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

func getRelatedMessages(cfg config.Config, logger logger.Logger, msg types.RequestMessage) ([]types.StoredMessage, error) {

	cc := clients.NewChromaClient(cfg)
	ids, err := cc.SearchForMessage(msg.Message, 5)
	if err != nil {
		logger.Error("Failed to search for related messages", err)
	}

	dc := store.NewDocumentStore() 

	messages, err := dc.FetchMessagesByIDs(ids)
	if err != nil {
		logger.Error("Failed to fetch messages from document store", err)
	}

	return messages, nil
}


func (c *ChatCapability) Tell(msg types.RequestMessage) {
	client := clients.NewChatClient(c.cfg.OpenAIKey)

	if msg.AuthUser.TelegramUserId < 0 {
		systemPrompt = fmt.Sprintf("%s\nFYI: You are in a group chat.", systemPrompt)
	}
	message := fmt.Sprintf("@%s (%s): %s", msg.Username, msg.Fullname, msg.Message)

	related, err := getRelatedMessages(*c.cfg, *c.logger, msg)
	if err != nil {
		c.logger.Error("Failed to get related messages", err)
	}

	c.logger.Info(">> Related messages", related)

	client.SetSystemPrompt(systemPrompt)
	client.AddMessage("system", "the next set of messages are unordered related messages returned from a vector search that may be relevant to the users query")
	client.AddStoredMessages(related)
	client.AddMessage("system", "the next set of messages are the last few messages in this chat in order")
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

func (c *ChatCapability) Describe() openai.Tool {
	fd := openai.FunctionDefinition{
		Name:        "ChatCapability",
		Description: "General chat capability that sends responses from gpt-4o",
	}
	return openai.Tool{
		Type:     openai.ToolTypeFunction,
		Function: &fd,
	}
}

func (c *ChatCapability) Check(inputMessage types.RequestMessage) float64 {
	return 0.0
}

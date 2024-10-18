package capabilities

import (
	_ "embed"

	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/pkg/types"
	o "github.com/sashabaranov/go-openai"
)

type TestCapability struct {
	logger *logger.Logger
	out    chan types.ResponseMessage
}

func NewTestCapability() *TestCapability {
	return &TestCapability{
		logger: logger.NewLogger("TestCapability"),
		out : make(chan types.ResponseMessage),
	}
}

func (t *TestCapability) SendMessage(msg types.RequestMessage) {
	t.logger.Info("Received message in test capability", msg)

	res := types.ResponseMessage {
		UserId: msg.UserId,
		ChatId: msg.ChatId,
		Message: "The user sent a test message so I will respond in kind",
	}

	t.out <- res
}

func (t *TestCapability) Check(msg types.RequestMessage) float64 {
	// if the message is the word test
	if msg.Message == "test" {
		return 1.0
	}
	return 0.0
}

func (t *TestCapability) Describe() o.Tool {
	fd := o.FunctionDefinition{
		Name:        "TestCapability",
		Description: "Capability for when the user sends only the word test",
	}

	return o.Tool{
		Type:     o.ToolTypeFunction,
		Function: &fd,
	}
}

func (i *TestCapability) GetUpdatesChan() chan types.ResponseMessage {
	return i.out
}

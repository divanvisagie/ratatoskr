package capabilities

import (
	_ "embed"

	"github.com/divanvisagie/ratatoskr/internal/config"
	"github.com/divanvisagie/ratatoskr/pkg/types"

	o "github.com/sashabaranov/go-openai"
)

type ImageGenerationCapability struct {
	cfg  *config.Config
	out  chan types.ResponseMessage
	done chan bool
}

// create constructor function
func NewImageGenerationCapability(cfg *config.Config) *ImageGenerationCapability {
	// load prompt string from file prompt.txt and include at build time

	instance := &ImageGenerationCapability{
		cfg:  cfg,
		out:  make(chan types.ResponseMessage),
		done: make(chan bool),
	}

	go types.ListenAndRespond(instance, instance.out)

	return instance
}

func (i *ImageGenerationCapability) SendMessage(msg types.RequestMessage) {
	i.out <- types.ResponseMessage{
		ChatId:  msg.ChatId,
		Message: "I can't generate images yet",
	}
}

func (i *ImageGenerationCapability) GetUpdatesChan() chan types.ResponseMessage {
	return i.out
}

func (i *ImageGenerationCapability) Describe() o.Tool{
	fd := o.FunctionDefinition{
		Name:        "ImageGenerationCapability",
		Description: "Can generate an image with dall-e if the user asks for it",
	}

	return o.Tool{
		Type: o.ToolTypeFunction,
		Function: &fd,
	}
}

package chat

import (
	_ "embed"

	"github.com/divanvisagie/ratatoskr/pkg/openai"
	"github.com/divanvisagie/ratatoskr/pkg/types"
)

type ChatCapability struct {
	client *openai.Client	
	in chan string
	out chan types.ResponseMessage
	done chan bool
}

//go:embed prompt.txt
var systemPrompt string

// create constructor function
func NewChatCapability() *ChatCapability {
	// load prompt string from file prompt.txt and include at build time

	return &ChatCapability{}
}

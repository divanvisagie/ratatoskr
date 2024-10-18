package types

import (
	openai "github.com/sashabaranov/go-openai"
)

/*
A cortex is an object that receives messages of type RequestMessage
and exposes a channel where you can subscribe to its outputs
*/
type Cortex interface {
	// Send a message to the cortex for processing
	SendMessage(RequestMessage)

	// Get access to a channel of response messages from the cortex
	GetUpdatesChan() chan ResponseMessage
}

/*
A capability is a cortex that exposes additional methods to assist with
selection
*/
type Capability interface {
	Cortex
	Describe() openai.Tool
	Check(inputMessage RequestMessage) float64
}

/*
ListenAndRespond listens for updates from an actor and sends them to
a channel
*/
func ListenAndRespond(actor Cortex, out chan ResponseMessage) {
	for {
		select {
		case response := <-actor.GetUpdatesChan():
			out <- response
		}
	}
}

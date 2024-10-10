package layers

import (
	"log"

	"github.com/divanvisagie/ratatoskr/pkg/types"
)

type MemoryLayer struct {
	data map[string][]string
	out  chan types.ResponseMessage
	next types.Cortex
}

func NewMemoryLayer(nextLayer types.Cortex) *MemoryLayer {
	instance := &MemoryLayer{
		data: make(map[string][]string),
		out:  make(chan types.ResponseMessage),
		next: nextLayer,
	}

	go types.ListenAndRespond(instance.next, instance.out)

	return instance
}

func (c *MemoryLayer) SendMessage(message types.RequestMessage) {
	// Add the request message to the map if it does not exist
	id := string(message.ChatId)
	if _, ok := c.data[id]; !ok {
		c.data[id] = []string{}
	}
	c.data[id] = append(c.data[id], message.Message)

	log.Println("Memory Layer: ", message)

	c.next.SendMessage(message)
}

func (c *MemoryLayer) GetUpdatesChan() chan types.ResponseMessage {
	return c.out
}


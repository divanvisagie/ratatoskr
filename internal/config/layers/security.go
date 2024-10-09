package layers

import (
	"log"

	"github.com/divanvisagie/ratatoskr/pkg/types"
)

type SecurityLayer struct {
	in   chan types.RequestMessage
	out  chan types.ResponseMessage
	done chan bool
}

func NewSecurityLayer() *SecurityLayer {
	layer := &SecurityLayer{
		in: make(chan types.RequestMessage),
		out: make(chan types.ResponseMessage),
		done: make(chan bool),
	}

	go layer.process()

	return layer
}

func (c *SecurityLayer) process() {
	for {
		select {
		case msg := <-c.in:
			log.Println("Security Layer: ", msg)
			// do something with the message
			responsemessage := types.ResponseMessage{
				ChatId: msg.ChatId,
				Content: "Hello, im the security layer",
			}
			c.out <- responsemessage
		case <-c.done:
			return
		}
	}
}

func (c *SecurityLayer) SendMessage(message types.RequestMessage) {
	c.in <- message
}

func (c *SecurityLayer) ReceiveMessage() types.ResponseMessage {
	return <-c.out
}

func (c *SecurityLayer) GetUpdatesChan() chan types.ResponseMessage{
	return c.out
}

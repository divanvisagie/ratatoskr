package layers

import (
	"log"

	"github.com/divanvisagie/ratatoskr/pkg/types"
)

type SecurityLayer struct {
	out  chan types.ResponseMessage
	next types.Cortex
}

func NewSecurityLayer(next types.Cortex) *SecurityLayer {
	layer := &SecurityLayer{
		out:  make(chan types.ResponseMessage),
		next: next,
	}

	go types.ListenAndRespond(layer.next, layer.out)

	return layer
}

func (s *SecurityLayer) SendMessage(message types.RequestMessage) {
	log.Println("Security Layer: ", message)
	s.next.SendMessage(message)
}

func (s *SecurityLayer) GetUpdatesChan() chan types.ResponseMessage {
	return s.out
}

func (s *SecurityLayer) Stop() {
	log.Println("Stopping Security Layer")
}

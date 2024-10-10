package layers

import (
	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/pkg/types"
)

type SecurityLayer struct {
	out    chan types.ResponseMessage
	next   types.Cortex
	logger *logger.Logger
}

func NewSecurityLayer(next types.Cortex) *SecurityLayer {
	layer := &SecurityLayer{
		out:    make(chan types.ResponseMessage),
		next:   next,
		logger: logger.NewLogger("SecurityLayer"),
	}

	go types.ListenAndRespond(layer.next, layer.out)

	return layer
}

func (s *SecurityLayer) SendMessage(message types.RequestMessage) {
	s.logger.Info("Sending message to security layer", message)
	s.next.SendMessage(message)
}

func (s *SecurityLayer) GetUpdatesChan() chan types.ResponseMessage {
	return s.out
}

func (s *SecurityLayer) Stop() {
	s.logger.Info("Stopping Security Layer")
}

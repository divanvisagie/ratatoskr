package layers

import (
	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/pkg/store"
	"github.com/divanvisagie/ratatoskr/pkg/types"
)

type SecurityLayer struct {
	out    chan types.ResponseMessage
	next   types.Cortex
	logger *logger.Logger
	store  *store.DocumentStore
}

func NewSecurityLayer(next types.Cortex) *SecurityLayer {
	layer := &SecurityLayer{
		out:    make(chan types.ResponseMessage),
		next:   next,
		logger: logger.NewLogger("SecurityLayer"),
		store:  store.NewDocumentStore(),
	}

	go types.ListenAndRespond(layer.next, layer.out)

	return layer
}

func (s *SecurityLayer) SendMessage(message types.RequestMessage) {
	s.logger.Info("Sending message to security layer", message)

	response := types.ResponseMessage{
		ChatId: message.ChatId,
		UserId: message.UserId,
	}

	// Check if the user is authorised
	user, err := s.store.GetUserByTelegramId(message.AuthUser.TelegramUserId)
	if err != nil {
		s.logger.Error("Failed to fetch user from memory layer", err)
		response.Message = "Error authorising user"
		s.out <- response
		return
	}

	if user == nil {
		s.logger.Info("User not found by id, trying username")
		user, err = s.store.GetUserByTelegramUsername(message.AuthUser.ChatName)
		if err != nil {
			s.logger.Error("Failed to fetch user from memory layer", err)
			response.Message = "Error authorising user"
			s.out <- response
			return
		}

		if user == nil {
			s.logger.Info("User not found in memory layer")
			response.Message = "User is not authorised to use this bot, contact @DivanVisagie"
			s.out <- response
			return
		}

		// Since the user was found by username, update the user's telegram id
		user.TelegramUserId = message.AuthUser.TelegramUserId

		err := s.store.SaveUser(*user)
		if err != nil {
			s.logger.Error("Failed to save user to memory layer", err)
			response.Message = "There was an error while trying to complete your signup"
			s.out <- response
			return
		}
	}

	s.next.SendMessage(message)
}

func (s *SecurityLayer) GetUpdatesChan() chan types.ResponseMessage {
	return s.out
}

func (s *SecurityLayer) Stop() {
	s.logger.Info("Stopping Security Layer")
}

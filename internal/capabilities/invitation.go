package capabilities

import (
	_ "embed"
	"fmt"

	"github.com/divanvisagie/ratatoskr/internal/logger"
	"github.com/divanvisagie/ratatoskr/pkg/store"
	"github.com/divanvisagie/ratatoskr/pkg/types"
	o "github.com/sashabaranov/go-openai"
)

type InvitationCapability struct {
	logger *logger.Logger
	out    chan types.ResponseMessage
}

func NewInvitationCapability() *InvitationCapability {
	return &InvitationCapability{
		logger: logger.NewLogger("InvitationCapability"),
		out:    make(chan types.ResponseMessage),
	}
}

func (i *InvitationCapability) Tell(msg types.RequestMessage) {
	i.logger.Info("Received message in invitation capability", msg)

	res := types.ResponseMessage{
		UserId:  msg.UserId,
		ChatId:  msg.ChatId,
		Message: "An error occurred while trying to invite the user",
	}

	// the message will be /invite @username
	// what we want is just the username without
	//the /invite @
	username := msg.Message[9:]

	i.logger.Info("The user wants to invite %s", username)

	ds := store.NewDocumentStore()

	// first see if there is already an invite
	user, err := ds.GetUserByTelegramUsername(username)

	if err != nil {
		i.logger.Error("Failed to fetch user from memory layer", err)
		res.Message = "Error authorising user"
	}

	if user != nil {
		res.Message = fmt.Sprintf("User %s has already been invited", username)
	} else {
		newUser := store.User{
			Username: username,
			Role:     "standard",
		}
		ds.SaveUser(newUser)
		res.Message = fmt.Sprintf("User %s has been invited", username)
	}

	i.out <- res
}

func (i *InvitationCapability) Check(msg types.RequestMessage) float64 {
	// if the message starts with /invite
	if len(msg.Message) >= 7 && msg.Message[:7] == "/invite" {
		return 1.0
	}
	return 0.0
}

func (i *InvitationCapability) GetUpdatesChan() chan types.ResponseMessage {
	return i.out
}

func (i *InvitationCapability) Describe() o.Tool {
	fd := o.FunctionDefinition{
		Name:        "InvitationCapability",
		Description: "Capability for messages that bgin with /invite",
	}

	return o.Tool{
		Type:     o.ToolTypeFunction,
		Function: &fd,
	}
}

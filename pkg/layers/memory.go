package layers

import (
	"fmt"
	"ratatoskr/pkg/repos"
	"ratatoskr/pkg/types"
)

type MemoryLayer struct {
	child Layer
	repo  *repos.MessageRepo
}

func NewMemoryLayer(repo *repos.MessageRepo, child Layer) *MemoryLayer {
	return &MemoryLayer{
		child,
		repo,
	}
}

func (m *MemoryLayer) getContext(username string) ([]types.StoredMessage, error) {
	// Get the context in order to feed future prompts
	history, err := m.repo.GetMessages(username)
	if err != nil {
		fmt.Printf("Error in memory layer when getting messages: %v\n", err)
		return nil, err
	}

	// Select only the latest 10 messages
	if len(history) > 10 {
		history = history[len(history)-10:]
	}
	fmt.Printf("History: %v\n", history)
	return history, nil
}

func (m *MemoryLayer) PassThrough(req *types.RequestMessage) (types.ResponseMessage, error) {
	inputMessage := repos.NewStoredMessage(repos.User, req.Message)
	m.repo.SaveMessage(req.UserName, *inputMessage)

	history, err := m.getContext(req.UserName)
	if err != nil {
		fmt.Printf("Error in memory layer when getting context: %v\n", err)
		return types.ResponseMessage{}, err
	}
	req.Context = history

	// Now pass through child layer
	res, err := m.child.PassThrough(req)
	if err != nil {
		fmt.Printf("Error in memory layer when passing through child layer: %v\n", err)
		return types.ResponseMessage{}, err
	}

	outputMessage := repos.NewStoredMessage(repos.Assistant, res.Message)
	m.repo.SaveMessage(req.UserName, *outputMessage)

	return res, nil
}

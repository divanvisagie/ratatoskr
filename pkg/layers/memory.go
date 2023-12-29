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

func (m *MemoryLayer) PassThrough(req *types.RequestMessage) (types.ResponseMessage, error) {
	inputMessage := repos.NewStoredMessage(repos.User, req.Message)
	m.repo.SaveMessage(req.UserName, *inputMessage)

	fmt.Printf("Saved inout message in memory layer: %v\n", req.Message)
	history, err := m.repo.GetMessages(req.UserName)
	if err != nil {
		fmt.Printf("Error in memory layer when getting messages: %v\n", err)
		return types.ResponseMessage{}, err
	}
	req.Context = history
	fmt.Printf("Memory layer got history: %+v\n", history)

	res, err := m.child.PassThrough(req)
	if err != nil {
		fmt.Printf("Error in memory layer when passing through child layer: %v\n", err)
		return types.ResponseMessage{}, err
	}

	outputMessage := repos.NewStoredMessage(repos.Assistant, res.Message)
	m.repo.SaveMessage(req.UserName, *outputMessage)

	return res, nil
}

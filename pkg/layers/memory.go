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
	m.repo.SaveMessage(repos.User, req.UserName, req.Message)
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

	m.repo.SaveMessage(repos.Assistant, req.UserName, res.Message)

	return res, nil
}

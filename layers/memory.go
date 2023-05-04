package layers

import (
	"ratatoskr/repositories"
	"ratatoskr/types"
)

type MemoryLayer struct {
	child Layer
	repo  *repositories.MessageRepository
}

func NewMemoryLayer(repo *repositories.MessageRepository, child Layer) *MemoryLayer {
	return &MemoryLayer{
		child,
		repo,
	}
}

func (m *MemoryLayer) PassThrough(req *types.RequestMessage) (types.ResponseMessage, error) {
	history := m.repo.GetMessages(req.UserName)
	req.Context = history

	res, err := m.child.PassThrough(req)
	if err != nil {
		return types.ResponseMessage{}, err
	}

	m.repo.SaveMessage(repositories.User, req.UserName, req.Message)
	m.repo.SaveMessage(repositories.Assistant, req.UserName, res.Message)

	return res, nil
}

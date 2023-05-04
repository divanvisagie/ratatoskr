package layers

import (
	"ratatoskr/repositories"
	"ratatoskr/types"
)

type MemoryLayer struct {
	child Layer
	repo  *repositories.Message
}

func NewMemoryLayer(repo *repositories.Message, child Layer) *MemoryLayer {
	return &MemoryLayer{
		child,
		repo,
	}
}

func (m *MemoryLayer) PassThrough(req *types.RequestMessage) (types.ResponseMessage, error) {
	history := m.repo.GetMessages(req.UserName)
	req.Context = history

	m.repo.SaveMessage(repositories.User, req.UserName, req.Message)

	res, err := m.child.PassThrough(req)
	if err != nil {
		return types.ResponseMessage{}, err
	}

	m.repo.SaveMessage(repositories.Assistant, req.UserName, res.Message)

	return res, nil
}

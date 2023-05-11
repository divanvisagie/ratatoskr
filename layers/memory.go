package layers

import (
	"ratatoskr/repos"
	"ratatoskr/types"
)

type MemoryLayer struct {
	child Layer
	repo  *repos.Message
}

func NewMemoryLayer(repo *repos.Message, child Layer) *MemoryLayer {
	return &MemoryLayer{
		child,
		repo,
	}
}

func (m *MemoryLayer) PassThrough(req *types.RequestMessage) (types.ResponseMessage, error) {
	history, err := m.repo.GetMessages(req.UserName)
	if err != nil {
		return types.ResponseMessage{}, err
	}
	req.Context = history

	m.repo.SaveMessage(repos.User, req.UserName, req.Message)

	res, err := m.child.PassThrough(req)
	if err != nil {
		return types.ResponseMessage{}, err
	}

	m.repo.SaveMessage(repos.Assistant, req.UserName, res.Message)

	return res, nil
}

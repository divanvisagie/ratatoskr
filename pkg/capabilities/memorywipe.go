package caps

import (
	"ratatoskr/pkg/repos"
	"ratatoskr/pkg/types"
)

type MemoryWipe struct {
	repo *repos.MessageRepo
}

func NewMemoryWipe(repo *repos.MessageRepo) *MemoryWipe {
	return &MemoryWipe{repo}
}

func (c MemoryWipe) Check(req *types.RequestMessage) float32 {
	if req.Message == "Clear memory" {
		return 1
	}
	return 0
}

func (c MemoryWipe) Execute(req *types.RequestMessage) (types.ResponseMessage, error) {
	c.repo.ClearMemory()
	res := types.ResponseMessage{
		ChatID:  req.ChatID,
		Message: "Memory wiped",
	}
	return res, nil
}

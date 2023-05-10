package caps

import (
	"ratatoskr/repos"
	"ratatoskr/types"
)

type MemoryWipe struct {
	repo *repos.Message
}

func NewMemoryWipe(repo *repos.Message) *MemoryWipe {
	return &MemoryWipe{repo}
}

func (c MemoryWipe) Check(req *types.RequestMessage) bool {
	return req.Message == "Clear memory"
}

func (c MemoryWipe) Execute(req *types.RequestMessage) (types.ResponseMessage, error) {
	c.repo.ClearMemory()
	res := types.ResponseMessage{
		ChatID:  req.ChatID,
		Message: "Memory wiped",
	}
	return res, nil
}

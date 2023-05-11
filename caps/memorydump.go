package caps

import (
	"fmt"
	"ratatoskr/repos"
	"ratatoskr/types"
	"strings"
)

type MemoryDump struct {
	repo *repos.Message
}

func NewMemoryDump(repo *repos.Message) *MemoryDump {
	return &MemoryDump{repo}
}

func (c MemoryDump) Check(req *types.RequestMessage) bool {
	return strings.ToLower(req.Message) == "memory dump"
}

func (c MemoryDump) Execute(req *types.RequestMessage) (types.ResponseMessage, error) {
	messages := c.repo.GetMessages(req.UserName)
	var message string
	for _, m := range messages {
		message += fmt.Sprintf("%s, %s\n", m.Role, m.Message)
	}
	res := types.ResponseMessage{
		ChatID:  req.ChatID,
		Message: message,
		Bytes:   []byte(message),
	}
	return res, nil
}

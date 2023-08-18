package caps

import (
	"fmt"
	"ratatoskr/pkg/repos"
	"ratatoskr/pkg/types"
	"strings"
)

type MemoryDump struct {
	repo *repos.Message
}

func NewMemoryDump(repo *repos.Message) *MemoryDump {
	return &MemoryDump{repo}
}

func (c MemoryDump) Check(req *types.RequestMessage) float32 {
	if strings.ToLower(req.Message) == "memory dump" {
		return 1
	}
	return 0
}

func (c MemoryDump) Execute(req *types.RequestMessage) (types.ResponseMessage, error) {
	messages, err := c.repo.GetMessages(req.UserName)
	if err != nil {
		return types.ResponseMessage{}, err
	}
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

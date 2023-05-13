package caps

import (
	"fmt"
	"ratatoskr/repos"
	"ratatoskr/types"
	"strings"
	"time"
)

type TestRedis struct {
	repo *repos.Message
}

func NewTestRedis(repo *repos.Message) *TestRedis {
	return &TestRedis{
		repo: repo,
	}
}

func (c TestRedis) Check(req *types.RequestMessage) float32 {
	if strings.ToLower(req.Message) == "test redis" {
		return 1
	}
	return 0
}

func (c TestRedis) Execute(req *types.RequestMessage) (types.ResponseMessage, error) {

	now := time.Now()
	timestamp := now.UnixMilli()
	key := fmt.Sprintf("%d", timestamp)

	c.repo.RememberEmbedded(repos.User, req.UserName, req.Message)

	res := types.ResponseMessage{
		ChatID:  req.ChatID,
		Message: fmt.Sprintf("Redis is working: %s", key),
	}
	return res, nil
}

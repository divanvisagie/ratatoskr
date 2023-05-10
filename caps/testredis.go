package caps

import (
	"context"
	"fmt"
	"os"
	"ratatoskr/types"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type TestRedis struct {
}

func NewTestRedis() *TestRedis {
	return &TestRedis{}
}

func (c TestRedis) Check(req *types.RequestMessage) bool {
	return strings.ToLower(req.Message) == "test redis"
}

func (c TestRedis) Execute(req *types.RequestMessage) (types.ResponseMessage, error) {
	ctx := context.Background()

	url := os.Getenv("REDIS_URL")
	opt, err := redis.ParseURL(url)
	if err != nil {
		res := types.ResponseMessage{
			ChatID:  req.ChatID,
			Message: "Redis is not working",
		}
		return res, nil
	}

	rdb := redis.NewClient(opt)
	now := time.Now()
	timestamp := now.UnixMilli()
	key := fmt.Sprintf("%d", timestamp)

	rdb.Set(ctx, key, "working", time.Millisecond*500)

	answer, err := rdb.Get(ctx, "messages").Result()
	if err != nil {
		res := types.ResponseMessage{
			ChatID:  req.ChatID,
			Message: "Redis broke on retreival",
		}
		return res, nil
	}

	res := types.ResponseMessage{
		ChatID:  req.ChatID,
		Message: fmt.Sprintf("Redis is working: %s", answer),
	}
	return res, nil
}

package layers

import (
	"fmt"
	"os"
	"ratatoskr/types"
)

type Security struct {
	//https://github.com/divanvisagie/Ratatoskr
	admin string
	child Layer
}

func NewSecurity(child Layer) *Security {
	//get from env variable
	admin := os.Getenv("TELEGRAM_ADMIN")

	return &Security{
		admin: admin,
	}
}

func (s *Security) PassThrough(req *types.RequestMessage) (types.ResponseMessage, error) {
	if req.UserName != s.admin {
		return types.ResponseMessage{
			Message: fmt.Sprintf(`
				Only %s is authorized to use this bot.
				However the code is open source and you can run your own instance of the bot.
				Just visit https://github.com/divanvisagie/Ratatoskr for more info. 
			`, s.admin),
			ChatID: req.ChatID,
		}, nil
	}

	res, err := s.child.PassThrough(req)
	return res, err
}

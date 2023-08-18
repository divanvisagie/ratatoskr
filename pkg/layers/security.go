package layers

import (
	"os"
	"ratatoskr/pkg/types"
	"strings"
)

type Security struct {
	//https://github.com/divanvisagie/Ratatoskr
	users []string
	child Layer
}

func NewSecurity(child Layer) *Security {
	//admin user
	admin := os.Getenv("TELEGRAM_ADMIN")
	//list of comma separated users
	users := os.Getenv("TELEGRAM_USERS")

	user_list := strings.Split(users, ",")

	//add admin user to list of users
	user_list = append(user_list, admin)

	return &Security{
		child: child,
		users: user_list,
	}
}

func (s *Security) PassThrough(req *types.RequestMessage) (types.ResponseMessage, error) {
	//check if user is in list of users
	authorized := false
	for _, user := range s.users {
		if user == req.UserName {
			authorized = true
			break
		}
	}

	if !authorized {
		return types.ResponseMessage{
			Message: strings.TrimSpace(`
				Only certain users are authorized to use this bot.
				However the code is open source and you can run your own instance of the bot.
				Just visit https://github.com/divanvisagie/Ratatoskr for more info. 
			`),
			ChatID: req.ChatID,
		}, nil
	}

	res, err := s.child.PassThrough(req)
	return res, err
}

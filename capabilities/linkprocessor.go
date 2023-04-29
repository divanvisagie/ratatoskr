package capabilities

import (
	"ratatoskr/types"
	"regexp"
)

type LinkProcessor struct {
}

func containsLink(message string) bool {
	//check if message contains a link
	r := regexp.MustCompile(`(http|https)://`)
	return r.MatchString(message)
}

func NewLinkProcessor() *LinkProcessor {
	return &LinkProcessor{}
}

func (c LinkProcessor) Check(req *types.RequestMessage) bool {
	return containsLink(req.Message)
}

func (c LinkProcessor) Execute(req *types.RequestMessage) (types.ResponseMessage, error) {
	//Extract link from message
	r := regexp.MustCompile(`(http|https)://[^\s]+`)
	link := r.FindString(req.Message)

	// Scrape website for main content
	res := types.ResponseMessage{
		ChatID:  req.ChatID,
		Message: link,
	}
	return res, nil
}

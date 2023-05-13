package caps

import (
	"fmt"
	client "ratatoskr/client"
	"ratatoskr/repos"
	"ratatoskr/types"
	"regexp"
)

type LinkProcessor struct {
	repo *repos.Message
}

func NewLinkProcessor(repo *repos.Message) *LinkProcessor {
	return &LinkProcessor{repo}
}

func (c LinkProcessor) Check(req *types.RequestMessage) bool {
	if !getFeatureIsEnabled("link") {
		return false
	}
	return containsLink(req.Message)
}

func (c LinkProcessor) Execute(req *types.RequestMessage) (types.ResponseMessage, error) {
	//Extract link from message
	r := regexp.MustCompile(`(http|https)://[^\s]+`)
	link := r.FindString(req.Message)

	body, err := client.ExtractBodyFromWebsite(link)
	if err != nil {
		return types.ResponseMessage{}, err
	}
	c.repo.SaveMessage(repos.System, req.UserName, fmt.Sprintf(`Website body text: %s`, body))

	context := getContextFromRepo(c.repo, req.UserName)

	summary := getSummaryFromChatGpt(context)

	// Scrape website for main content
	res := types.ResponseMessage{
		ChatID:  req.ChatID,
		Message: summary,
	}
	return res, nil
}

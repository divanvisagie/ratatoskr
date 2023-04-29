package capabilities

import (
	"os"
	client "ratatoskr/clients"
	"ratatoskr/types"
	"regexp"
	"strings"
)

type Notion struct {
	admin        string
	systemPrompt string
}

func NewNotion() *Notion {
	admin := os.Getenv("TELEGRAM_ADMIN")
	systemPrompt := strings.TrimSpace(`Ratatoskr, is an EI (Extended Intelligence). 
	An extended intelligence is a software system 
	that utilises multiple Language Models, AI models, 
	NLP Functions and other capabilities to best serve 
	the user.

	You are part of the link processing module, whose job it is to take a
	link and return a summary of the content that will provide good keywords
	when searching for it since the link and the summary will be stored in
	Notion wof the user. You will be provided the link 
	and a summary message if possible. The summary is extracted 
	directly from the html body and may contain some junk data. If the page cannot be
	parsed the body will just be the value "None". Use the 
	body and your existing knowledge of the site where possible 
	to provide the best summary possible. Tell the user what the link is 
	about and what can be learned from it. Remember to highlight any stand 
	out points that may contain unexpected conclusions or information.`)
	return &Notion{admin, systemPrompt}
}

func (c Notion) Check(req *types.RequestMessage) bool {
	if req.UserName != c.admin {
		return false
	}

	r := regexp.MustCompile(`(http|https)://`)
	return r.MatchString(req.Message)
}

func (c Notion) Execute(req *types.RequestMessage) (types.ResponseMessage, error) {
	r := regexp.MustCompile(`(http|https)://[^\s]+`)
	link := r.FindString(req.Message)

	body, err := client.ExtractBodyFromWebsite(link)
	if err != nil {
		return types.ResponseMessage{}, err
	}

	summary := client.NewOpenAIClient(c.systemPrompt).Complete(body)
	// TODO: Write to Notion

	// Scrape website for main content
	res := types.ResponseMessage{
		ChatID:  req.ChatID,
		Message: summary,
	}
	return res, nil
}

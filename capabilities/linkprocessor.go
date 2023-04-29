package capabilities

import (
	"fmt"
	client "ratatoskr/clients"
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

func getSummaryFromChatGpt(link string, message string) string {
	systemPrompt := fmt.Sprintf(`Ratatoskr, 
		is an EI (Extended Intelligence). 
		An extended intelligence is a software system 
		that utilises multiple Language Models, AI models, 
		NLP Functions and other capabilities to best serve 
		the user.

		You are part of the link processing module, whose job it is to take a link and return a summary of the content.
		you will be provided the link and a summary message from the user as input. The summary is extracted directly
		from the html body and may contain some junk data. Use the summary message and your existing knowledge of the 
		site where possible to provide the best summary possible. Tell the user what the link is about and what can be learned from it.
		Remember to highlight any stand out points that may contain unexpected conclusions or information.

		link: %s
	`, link)

	summary := client.NewOpenAIClient(systemPrompt).Complete(message)
	return summary
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

	body, err := client.ExtractBodyFromWebsite(link)
	if err != nil {
		return types.ResponseMessage{}, err
	}

	summary := getSummaryFromChatGpt(link, body)

	// Scrape website for main content
	res := types.ResponseMessage{
		ChatID:  req.ChatID,
		Message: summary,
	}
	return res, nil
}

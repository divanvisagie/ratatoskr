package caps

import (
	"fmt"
	"os"
	"ratatoskr/pkg/client"
	"ratatoskr/pkg/repos"
	"ratatoskr/pkg/types"
	"regexp"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type Notion struct {
	admin        string
	systemPrompt string
	notion       *client.Notion
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
	Notion for the user. You will be provided the link by the user and the body will
	be provided in a system message just before it.

	The body is extracted directly from the html body of the website by the system
	and may contain some junk data. If the page cannot be
	parsed the body will just be the value "<None>". Use the 
	body and your existing knowledge of the site where possible 
	to provide the best summary possible. Tell the user what the link is 
	about and what can be learned from it. Remember to highlight any stand 
	out points that may contain unexpected conclusions or information.`)

	notion := client.NewNotion()

	return &Notion{admin, systemPrompt, notion}
}

func (c Notion) Check(req *types.RequestMessage) float32 {

	if !getFeatureIsEnabled("notion") {
		return 1
	}
	if req.UserName != c.admin {
		return 1
	}
	if containsLink(req.Message) {
		return 1
	}
	return 0
}

func (c Notion) Execute(req *types.RequestMessage) (types.ResponseMessage, error) {
	r := regexp.MustCompile(`(http|https)://[^\s]+`)
	link := r.FindString(req.Message)

	context := []types.StoredMessage{}

	//msg
	um := repos.NewStoredMessage(
		repos.User,
		req.Message,
	)
	context = append(context, *um)

	body, err := client.ExtractBodyFromWebsite(link)
	if err != nil {
		return types.ResponseMessage{}, err
	}

	sm := repos.NewStoredMessage(
		repos.System,
		fmt.Sprintf(`Website body text: %s`, body),
	)
	context = append(context, *sm)

	//get context
	history := make([]openai.ChatCompletionMessage, len(context))
	for i, message := range context {
		history[i] = messageToChatCompletionMessage(message)
	}

	summary := getSummaryFromChatGpt(history)

	result, err := c.notion.AddLinkToTodaysPage(link, summary)
	if err != nil {
		return types.ResponseMessage{}, err
	}

	responseText := fmt.Sprintf(`**I have added the following [here](%s) to your journal in Notion:**
	
	%s
	
	%s`, result.URL, summary, link)

	res := types.ResponseMessage{
		ChatID:  req.ChatID,
		Message: strings.TrimSpace(responseText),
		Context: context,
	}
	return res, nil
}

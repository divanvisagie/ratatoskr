package caps

import (
	client "ratatoskr/client"
	"ratatoskr/repos"
	"regexp"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

func getFeatureIsEnabled(feature string) bool {
	feaureMap := map[string]bool{
		"notion":  true,
		"link":    true,
		"chatGpt": true,
	}
	return feaureMap[feature]
}

func getSummaryFromChatGpt(context []openai.ChatCompletionMessage) string {
	systemPrompt := strings.TrimSpace(`Ratatoskr, is an EI (Extended Intelligence). 
	An extended intelligence is a software system 
	that utilises multiple Language Models, AI models, 
	NLP Functions and other capabilities to best serve 
	the user.

	You are part of the link processing module, whose job it is to take a
	link and return a summary of the content that will provide good keywords
	when searching for it since the link and the summary may be stored in
	Notion for the user. You will be provided the link and possibly a query by the user 
	the system will provide the body of the page in a system message once it has looked it up.

	The body is extracted directly from the html body of the website by the system
	and may contain some junk data. If the page cannot be
	parsed the body will just be the value "<None>". Use the 
	body returned by the system and your existing knowledge of the site where possible 
	to provide the best summary possible. Tell the user what the link is 
	about and what can be learned from it. Remember to highlight any stand 
	out points that may contain unexpected conclusions or information.`)

	summary := client.NewOpenAIClient().AddSystemMessage(systemPrompt).SetHistory(context).Complete()
	return summary
}

func containsLink(message string) bool {
	//check if message contains a link
	r := regexp.MustCompile(`(http|https)://`)
	return r.MatchString(message)
}

func getContextFromRepo(repo *repos.Message, username string) []openai.ChatCompletionMessage {
	context, err := repo.GetMessages(username)
	if err != nil {
		//TODO: Handle this error better
		return []openai.ChatCompletionMessage{}
	}
	history := make([]openai.ChatCompletionMessage, len(context))
	for i, message := range context {
		history[i] = messageToChatCompletionMessage(message)
	}

	return history
}

package utils

import (
	"fmt"
	"log"

	"github.com/sashabaranov/go-openai"
	"github.com/tiktoken-go/tokenizer"
)

const MODEL_LIMIT = 4096

func Tokenize(text string) ([]string, error) {
	enc, err := tokenizer.ForModel(tokenizer.Model(tokenizer.GPT35Turbo))
	if err != nil {
		log.Printf("Error while creating tokenizer: %s", err)
	}

	// this should print a list of token ids
	_, words, _ := enc.Encode(text)
	fmt.Println(words)

	return words, nil
}

func ShortenContext(context []openai.ChatCompletionMessage, limit int) ([]openai.ChatCompletionMessage, error) {
	length := 0

	newContext := []openai.ChatCompletionMessage{}
	for i := (len(context) - 1); i >= 0; i-- {
		t, err := Tokenize(context[i].Content)
		if err != nil {
			log.Printf(`Error while tokenizing "%s": %s`, context[i].Content, err)
			return context, err
		}
		length += len(t)

		if length < limit {
			newContext = append(newContext, context[i])
		} else {
			break
		}
	}

	if len(newContext) < 2 {
		return context, fmt.Errorf("the context could not be shortened to %d tokens", limit)
	}

	return reverse(newContext), nil
}

func reverse[T any](a []T) []T {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
	return a
}

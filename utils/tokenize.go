package utils

import (
	"fmt"
	"strings"

	"github.com/tiktoken-go/tokenizer"
)

func Tokenize(text string) ([]string, error) {
	words := strings.Fields(text)

	enc, err := tokenizer.ForModel(tokenizer.Model(tokenizer.GPT35Turbo))
	if err != nil {
		panic("oh oh")
	}

	// this should print a list of token ids
	ids, _, _ := enc.Encode(text)
	fmt.Println(ids)

	return words, nil
}

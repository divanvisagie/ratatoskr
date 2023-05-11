package utils

import (
	"testing"

	"github.com/sashabaranov/go-openai"
)

func TestTokenize(t *testing.T) {
	actual, err := Tokenize("this is a test")

	if len(actual) != 4 {
		t.Errorf("Expected 4 tokens, got %d", len(actual))
	}

	if err != nil {
		t.Errorf("Error while tokenizing: %s", err)
	}
}

func TestShortenContext(t *testing.T) {
	actual, err := ShortenContext([]openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: "Hello, I am a chatbot.",
		},
		{
			Role:    "user",
			Content: "How are you?",
		},
		{
			Role:    "assistant",
			Content: "I am fine, thank you.",
		},
		{
			Role:    "user",
			Content: "What is your name?",
		},
	}, 13)

	if err != nil {
		t.Errorf("Error while shortening context: %s", err)
	}

	if len(actual) != 3 {
		t.Errorf("Expected 2 messages, got %d", len(actual))
	}

	// expect last message to be the same
	if actual[2].Content != "What is your name?" {
		t.Errorf("Expected last message to be 'What is your name?', got %s", actual[2].Content)
	}
}

func TestShortenContextWith1Item(t *testing.T) {
	actual, err := ShortenContext([]openai.ChatCompletionMessage{
		{
			Role:    "system",
			Content: "Hello, I am a chatbot.",
		},
		{
			Role:    "user",
			Content: "test",
		},
	}, 4096)

	if err != nil {
		t.Errorf("Expected error, got nil")
	}

	if len(actual) != 2 {
		t.Errorf("Expected 1 message, got %d", len(actual))
	}
}

func TestReverse(t *testing.T) {
	actual := reverse([]string{"a", "b", "c"})

	if actual[0] != "c" {
		t.Errorf("Expected first element to be 'c', got %s", actual[0])
	}

	if actual[1] != "b" {
		t.Errorf("Expected second element to be 'b', got %s", actual[1])
	}

	if actual[2] != "a" {
		t.Errorf("Expected third element to be 'a', got %s", actual[2])
	}

	if len(actual) != 3 {
		t.Errorf("Expected 3 tokens, got %d", len(actual))
	}
}

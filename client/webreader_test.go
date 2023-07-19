package client

import (
	"strings"
	"testing"
)

func TestShorten(t *testing.T) {
	actual := shorten(`this is a text that contains many many words
	in fact, it contains so many words that it is hard to count them all
	so we will just assume that there are many words`, 3)

	expected := `this is a`
	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

func TestExtractBodyFromWebsite(t *testing.T) {
	urls := []string{
		"https://github.com/ThePrimeagen/aoc/blob/2022/src/bin/day6_2.rs",
		"https://divanv.com/post/service-registry/",
	}

	for _, url := range urls {
		bodyText, err := ExtractBodyFromWebsite(url)
		if err != nil {
			t.Errorf("Error while extracting body from website: %s", err)
		}

		if len(bodyText) == 0 {
			t.Errorf("Expected body text to be longer than 0 characters")
		}

		// Check if the extracted text does not contain script-related content
		if strings.Contains(strings.ToLower(bodyText), "<script>") ||
			strings.Contains(strings.ToLower(bodyText), "</script>") ||
			strings.Contains(strings.ToLower(bodyText), "<noscript>") ||
			strings.Contains(strings.ToLower(bodyText), "</noscript>") {
			t.Errorf("Extracted text should not contain script-related content")
		}
	}
}

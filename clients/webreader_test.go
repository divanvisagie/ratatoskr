package client

import "testing"

func TestShorten(t *testing.T) {
	actual := shorten(`this is a text that contains many many words
	in fact, it contains so many words that it is hard to count them all
	so we will just assume that there are many words`, 3)

	expected := `thi`
	if actual != expected {
		t.Errorf("Expected %s, got %s", expected, actual)
	}
}

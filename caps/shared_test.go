package caps

import "testing"

// test

func TestContainsLink(t *testing.T) {
	ti := []struct {
		input    string
		expected bool
	}{
		{"this is a message that contains a link https://www.google.com", true},
		{"this is a message that contains a link http://www.google.com", true},
		{"this is a message that contains a link www.google.com", false},
		{"test", false},
		{"this is a message that contains a link google.com/abc", false},
	}

	for _, x := range ti {

		actual := containsLink(x.input)
		expected := x.expected
		if actual != expected {
			t.Errorf("Expected %t, got %t", expected, actual)
		}
	}
}

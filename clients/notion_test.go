package client

import (
	"fmt"
	"testing"
)

func TestCreatePageFprToday(t *testing.T) {
	tags := []string{"Ratatoskr", "UNIT_TEST"}
	n := NewNotion()
	page, err := n.CreatePageForToday(tags)

	fmt.Println(page)
	if err != nil {
		t.Errorf("Error while creating page: %s", err)
	}
}

func TestWriteToPage(t *testing.T) {

	WriteToPage("https://www.google.com", "This is a test")
}

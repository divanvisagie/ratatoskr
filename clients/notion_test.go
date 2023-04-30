package client

import (
	"fmt"
	"testing"
)

func TestGetTodaysPage(t *testing.T) {
	c := NewNotion()
	page, err := c.GetTodaysPage()

	fmt.Println(page)

	if err != nil {
		t.Errorf("Error while getting today's page: %s", err)
	}
}

func TestAddLinkToTodaysPage(t *testing.T) {
	c := NewNotion()
	err := c.AddLinkToTodaysPage("https://www.youtube.com/watch?v=dQw4w9WgXcQ", "This is a test link")

	if err != nil {
		t.Errorf("Error while adding link to today's page: %s", err)
	}
}

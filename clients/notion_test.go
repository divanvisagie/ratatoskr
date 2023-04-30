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

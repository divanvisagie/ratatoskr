package client

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/dstotijn/go-notion"
)

func WriteToPage(link string, summary string) {
	token := os.Getenv("NOTION_TOKEN")
	journalPageID := os.Getenv("NOTION_JOURNAL_DB")

	fmt.Print(journalPageID)
	client := notion.NewClient(token)

	page, err := client.FindPageByID(context.Background(), "18d35eb5-91f1-4dcb-85b0-c340fd965015")
	if err != nil {
		log.Println(err)
	}

	fmt.Print(page)
}

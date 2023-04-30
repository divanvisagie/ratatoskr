package client

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	notion "github.com/dstotijn/go-notion"
)

type Notion struct {
	client        *notion.Client
	journalPageID string
}

func NewNotion() *Notion {
	token := os.Getenv("NOTION_TOKEN")
	client := notion.NewClient(token)
	journalPageID := os.Getenv("NOTION_JOURNAL_DB")

	return &Notion{client, journalPageID}
}

func (n *Notion) CreatePageForToday(tags []string) (notion.Page, error) {
	//get todays date in the format YYYY-MM-DD
	date := time.Now().Format("2006-01-02")
	today, err := notion.ParseDateTime(date)
	if err != nil {
		log.Println(err)
	}

	tagOptions := make([]notion.SelectOptions, len(tags))
	//tags to SelectOptions
	for i, tag := range tags {
		so := notion.SelectOptions{
			Name: tag,
		}
		tagOptions[i] = so
	}

	titleRichText := notion.RichText{
		Text: &notion.Text{
			Content: "Test Page",
		},
	}
	titleRTC := []notion.RichText{titleRichText}

	databasePageProperties := &notion.DatabasePageProperties{
		"Date": notion.DatabasePageProperty{
			Date: &notion.Date{
				Start: today,
			},
		},
		"Tags": notion.DatabasePageProperty{
			MultiSelect: tagOptions,
		},
	}

	createPageParams := notion.CreatePageParams{
		ParentType:             notion.ParentTypeDatabase,
		ParentID:               n.journalPageID,
		DatabasePageProperties: databasePageProperties,
		Title:                  titleRTC,
	}

	//create a new page with the date as the title
	page, err := n.client.CreatePage(context.Background(), createPageParams)
	if err != nil {
		log.Println(err)
		return notion.Page{}, err
	}
	return page, nil
}

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

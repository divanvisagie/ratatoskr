package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

type Notion struct {
	journalPageID string
	token         string
}

type ResultObject struct {
	Object      string `json:"object"`
	ID          string `json:"id"`
	CreatedTime string `json:"created_time"`
}

type TodaysPageResult struct {
	Object  string         `json:"object"`
	Results []ResultObject `json:"results"`
}

func NewNotion() *Notion {
	token := os.Getenv("NOTION_TOKEN")
	journalPageID := os.Getenv("NOTION_JOURNAL_DB")

	return &Notion{journalPageID, token}
}

func (n *Notion) GetTodaysPage() (TodaysPageResult, error) {
	today := time.Now().Format("2006-01-02")

	requestBody := map[string]interface{}{
		"filter": map[string]interface{}{
			"property": "Date",
			"date": map[string]interface{}{
				"on_or_after": today,
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Println(err)
		return TodaysPageResult{}, err
	}

	// Create a new HTTP request with the given method and URL.
	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("https://api.notion.com/v1/databases/%s/query",
			n.journalPageID,
		),
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		log.Println(err)
		return TodaysPageResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", n.token))
	req.Header.Set("Notion-Version", "2022-06-28")

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return TodaysPageResult{}, err
	}
	defer resp.Body.Close()

	// Read the response body.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return TodaysPageResult{}, err
	}

	// Parse the response JSON into a Response struct
	var response TodaysPageResult
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error unmarshaling response body:", err)
		return TodaysPageResult{}, err
	}
	return response, nil
}

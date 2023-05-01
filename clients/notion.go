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

func doRequestToNotion(method string, token string, url string, body []byte) ([]byte, error) {
	req, err := http.NewRequest(
		method,
		url,
		bytes.NewBuffer(body),
	)
	if err != nil {
		log.Println(err)
		return make([]byte, 0), err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	req.Header.Set("Notion-Version", "2022-06-28")

	// Make the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err)
		return make([]byte, 0), err
	}
	defer resp.Body.Close()

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return make([]byte, 0), err
	}

	return body, nil
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

	url := fmt.Sprintf("https://api.notion.com/v1/databases/%s/query", n.journalPageID)
	body, err := doRequestToNotion("POST", n.token, url, jsonBody)
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

func (n *Notion) AddLinkToTodaysPage(link string, summary string) error {

	todaysPage, err := n.GetTodaysPage()
	if err != nil {
		log.Println(err)
		return err
	}

	var result ResultObject
	if len(todaysPage.Results) == 0 {
		r, err := n.CreatePageForToday([]string{"Ratatoskr"})
		if err != nil {
			log.Println(err)
			return err
		}
		result = r
	} else {
		result = todaysPage.Results[0]
	}

	requestBody := map[string]interface{}{
		"children": []interface{}{
			map[string]interface{}{
				"object": "block",
				"type":   "paragraph",
				"paragraph": map[string]interface{}{
					"rich_text": []interface{}{
						map[string]interface{}{
							"type": "text",
							"text": map[string]interface{}{
								"content": summary,
							},
						},
					},
				},
			},
			map[string]interface{}{
				"object": "block",
				"type":   "bookmark",
				"bookmark": map[string]interface{}{
					"url": link,
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Println(err)
		return err
	}

	url := fmt.Sprintf("https://api.notion.com/v1/blocks/%s/children", result.ID)
	body, err := doRequestToNotion("PATCH", n.token, url, jsonBody)
	if err != nil {
		log.Println(err)
		return err
	}

	fmt.Println(string(body))

	return nil
}

func (n *Notion) CreatePageForToday(tags []string) (ResultObject, error) {
	dayOfWeek := time.Now().Format("Monday")
	today := time.Now().Format("2006-01-02")

	tag_options := make([]interface{}, len(tags))
	for i, tag := range tags {
		tag_options[i] = map[string]interface{}{
			"name": tag,
		}
	}

	requestBody := map[string]interface{}{
		"parent": map[string]interface{}{
			"database_id": n.journalPageID,
		},
		"properties": map[string]interface{}{
			"Tags": map[string]interface{}{
				"multi_select": tag_options,
			},
			"Date": map[string]interface{}{
				"date": map[string]interface{}{
					"start": today,
				},
			},
			"Name": map[string]interface{}{
				"title": []interface{}{
					map[string]interface{}{
						"text": map[string]interface{}{
							"content": dayOfWeek,
						},
					},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		log.Println(err)
		return ResultObject{}, err
	}

	url := "https://api.notion.com/v1/pages"
	body, err := doRequestToNotion("POST", n.token, url, jsonBody)
	if err != nil {
		log.Println(err)
		return ResultObject{}, err
	}

	// Parse the response JSON into a Response struct
	var response ResultObject
	err = json.Unmarshal(body, &response)
	if err != nil {
		fmt.Println("Error unmarshaling response body:", err)
		return ResultObject{}, err
	}

	return response, nil
}

package client

import (
	"fmt"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

func shorten(text string, limit int) string {

	//limit text to 1000 characters
	if len(text) > limit {
		text = text[:limit]
	}
	return text
}

func trimText(text string) string {
	// trimmed := strings.TrimSpace(text)
	trimmed := strings.ReplaceAll(text, "\n", "")

	trimmed = shorten(trimmed, 1000)

	return trimmed
}

func ExtractBodyFromWebsite(url string) (string, error) {
	// Sync code since this is intented to be a single user system
	wg := new(sync.WaitGroup)
	wg.Add(1)

	c := colly.NewCollector()
	bodyText := ""
	c.OnHTML("body", func(e *colly.HTMLElement) {

		bodyText = e.Text
		e.DOM.Find("#content").Each(func(i int, s *goquery.Selection) {
			fmt.Println(s.Text())
			bodyText = s.Text()
		})

		fmt.Println(e.Text)
		wg.Done()
	})
	c.Visit(url)

	wg.Wait()

	return trimText(bodyText), nil
}

// func ProcessLink(url string) (string, error) {
// 	// Make an HTTP GET request to the URL
// 	resp, err := http.Get(url)
// 	if err != nil {
// 		return "", err
// 	}
// 	defer resp.Body.Close()

// 	// Read the response body
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", err
// 	}

// 	// Extract the relevant text from the response body
// 	text := extractText(body)

// 	return text, nil
// }

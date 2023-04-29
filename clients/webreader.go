package client

import (
	"fmt"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
)

// Shorten text to a specific amount of tokens
func shorten(text string, limit int) string {
	words := strings.Fields(text)

	if len(words) > limit {
		// Split the input string into words using the Fields function from the strings package

		// Return the length of the resulting slice
		cut := words[:limit]
		return strings.Join(cut, " ")
	}

	return strings.Join(words, " ")
}

func trimText(text string) string {
	trimmed := strings.TrimSpace(text)
	trimmed = shorten(trimmed, 1000)

	return trimmed
}

func ExtractBodyFromWebsite(url string) (string, error) {
	// Sync code since this is intented to be a single user system
	wg := new(sync.WaitGroup)
	wg.Add(1)

	c := colly.NewCollector()

	bodyText := "None"
	c.OnHTML("body", func(e *colly.HTMLElement) {

		bodyText = e.Text
		e.DOM.Find("#content").Each(func(i int, s *goquery.Selection) {
			fmt.Println(s.Text())
			bodyText = s.Text()
		})

		fmt.Println(bodyText)
		wg.Done()
	})
	c.OnError(func(r *colly.Response, err error) {
		fmt.Println("Request URL:", r.Request.URL, "failed with response:", r, "\nError:", err)
		wg.Done()
	})
	c.Visit(url)

	wg.Wait()

	return trimText(bodyText), nil
}

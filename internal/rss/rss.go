package rss

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Items       []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func FetchFeed(ctx *context.Context, feedURL string) (*RSSFeed, error) {
	request, err := http.NewRequestWithContext(*ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("could not create the HTTP request: %w", err)
	}
	// Identify the program to the server
	request.Header.Set("User-Agent", "rssgator")

	client := http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}

	data, err := io.ReadAll(response.Body)
	defer response.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("could not read response body: %w", err)
	}

	// Extra XML fields are silently discarded, missing XML fields are stored as zero values
	rssFeed := &RSSFeed{}
	err = xml.Unmarshal(data, rssFeed)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal received data to XML: %w", err)
	}

	// Convert HTML entities into the equivalent characters
	rssFeed.Channel.Description = html.UnescapeString(rssFeed.Channel.Description)
	rssFeed.Channel.Title = html.UnescapeString(rssFeed.Channel.Title)
	for i, item := range rssFeed.Channel.Items {
		// We cannot assign `item` directly because it's a **copy**.
		rssFeed.Channel.Items[i].Description = html.UnescapeString(item.Description)
		rssFeed.Channel.Items[i].Title = html.UnescapeString(item.Title)
	}

	return rssFeed, nil
}

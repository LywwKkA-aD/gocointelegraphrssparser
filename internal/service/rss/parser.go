// internal/service/rss/parser.go
package rss

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/LywwKkA-aD/gocointelegraphrssparser/internal/models"
	"github.com/LywwKkA-aD/gocointelegraphrssparser/pkg/logger"
)

type Parser struct {
	client       *http.Client
	lastNewsTime time.Time
	firstFetch   bool
	mu           sync.RWMutex
}

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

type Channel struct {
	Items []Item `xml:"item"`
}

type Item struct {
	Title       string   `xml:"title"`
	Link        string   `xml:"link"`
	PubDate     string   `xml:"pubDate"`
	Description string   `xml:"description"`
	Categories  []string `xml:"category"`
	GUID        string   `xml:"guid"`
}

func NewParser() *Parser {
	return &Parser{
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		firstFetch: true,
	}
}

func (p *Parser) FetchNews(ctx context.Context, feedURL string) ([]models.NewsItem, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var feed RSS
	if err := xml.Unmarshal(body, &feed); err != nil {
		return nil, fmt.Errorf("parse XML: %w", err)
	}

	return p.processItems(feed.Channel.Items)
}

func (p *Parser) processItems(items []Item) ([]models.NewsItem, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var newsItems []models.NewsItem

	// Process only the latest item if it's the first fetch
	if p.firstFetch && len(items) > 0 {
		latestItem := items[0] // First item is the newest
		pubDate, err := time.Parse(time.RFC1123Z, strings.TrimSpace(latestItem.PubDate))
		if err != nil {
			return nil, fmt.Errorf("parse date for latest item: %w", err)
		}

		p.lastNewsTime = pubDate
		p.firstFetch = false

		newsItem := p.convertItemToNews(latestItem, pubDate)
		return []models.NewsItem{newsItem}, nil
	}

	// For subsequent fetches, check for newer items
	for _, item := range items {
		pubDate, err := time.Parse(time.RFC1123Z, strings.TrimSpace(item.PubDate))
		if err != nil {
			logger.Error("Failed to parse date %s: %v", item.PubDate, err)
			continue
		}

		// Break if we hit an old item
		if !pubDate.After(p.lastNewsTime) {
			break
		}

		newsItem := p.convertItemToNews(item, pubDate)
		newsItems = append(newsItems, newsItem)
	}

	// Update last news time if we found new items
	if len(newsItems) > 0 {
		p.lastNewsTime = newsItems[0].PubDate
	}

	return newsItems, nil
}

func (p *Parser) convertItemToNews(item Item, pubDate time.Time) models.NewsItem {
	return models.NewsItem{
		Title:       cleanCDATA(item.Title),
		Link:        item.GUID,
		PubDate:     pubDate,
		Description: extractLastParagraph(item.Description),
		Categories:  cleanCategories(item.Categories),
	}
}

func cleanCDATA(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "<![CDATA[")
	s = strings.TrimSuffix(s, "]]>")
	return strings.TrimSpace(s)
}

func extractLastParagraph(description string) string {
	startIdx := strings.LastIndex(description, "<p>")
	if startIdx == -1 {
		return cleanCDATA(description)
	}

	endIdx := strings.Index(description[startIdx:], "</p>")
	if endIdx == -1 {
		return cleanCDATA(description[startIdx+3:])
	}

	text := description[startIdx+3 : startIdx+endIdx]
	return strings.TrimSpace(text)
}

func cleanCategories(categories []string) []string {
	var cleaned []string
	for _, category := range categories {
		clean := cleanCDATA(category)
		if clean != "" {
			cleaned = append(cleaned, clean)
		}
	}
	return cleaned
}

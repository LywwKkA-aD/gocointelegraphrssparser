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
	logger.Debug("Fetching news from %s", feedURL)

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

	if p.firstFetch && len(items) > 0 {
		latestItem := items[0]
		pubDate, err := time.Parse(time.RFC1123Z, strings.TrimSpace(latestItem.PubDate))
		if err != nil {
			return nil, fmt.Errorf("parse date for latest item: %w", err)
		}

		logger.Info("First fetch: Latest article from %v", pubDate)
		p.lastNewsTime = pubDate
		p.firstFetch = false

		newsItem := p.convertItemToNews(latestItem, pubDate)
		return []models.NewsItem{newsItem}, nil
	}

	for _, item := range items {
		pubDate, err := time.Parse(time.RFC1123Z, strings.TrimSpace(item.PubDate))
		if err != nil {
			logger.Error("Failed to parse date %s: %v", item.PubDate, err)
			continue
		}

		if !pubDate.After(p.lastNewsTime) {
			logger.Debug("No newer articles found after %v", p.lastNewsTime)
			break
		}

		logger.Info("Found new article from %v", pubDate)
		newsItem := p.convertItemToNews(item, pubDate)
		newsItems = append(newsItems, newsItem)
	}

	if len(newsItems) > 0 {
		p.lastNewsTime = newsItems[0].PubDate
		logger.Info("Updated last news time to %v", p.lastNewsTime)
	}

	return newsItems, nil
}

func (p *Parser) convertItemToNews(item Item, pubDate time.Time) models.NewsItem {
	return models.NewsItem{
		Title:       cleanCDATA(item.Title),
		Link:        item.GUID,
		PubDate:     pubDate,
		Description: extractDescription(item.Description),
		Categories:  cleanCategories(item.Categories),
	}
}

func cleanCDATA(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "<![CDATA[")
	s = strings.TrimSuffix(s, "]]>")
	return strings.TrimSpace(s)
}

func extractDescription(description string) string {
	// Remove any HTML styling
	description = strings.ReplaceAll(description, `style="float:right; margin:0 0 10px 15px; width:240px;"`, "")

	// Remove image tags
	imgStart := strings.Index(description, "<img")
	if imgStart != -1 {
		imgEnd := strings.Index(description[imgStart:], ">")
		if imgEnd != -1 {
			description = description[:imgStart] + description[imgStart+imgEnd+1:]
		}
	}

	// Extract text from p tags
	pStart := strings.Index(description, "<p>")
	if pStart != -1 {
		pEnd := strings.Index(description[pStart:], "</p>")
		if pEnd != -1 {
			description = description[pStart+3 : pStart+pEnd]
		}
	}

	// Clean CDATA and trim
	description = cleanCDATA(description)

	// Replace problematic characters
	description = strings.ReplaceAll(description, `“`, `"`)
	description = strings.ReplaceAll(description, `”`, `"`)
	description = strings.ReplaceAll(description, "'", "'")
	description = strings.ReplaceAll(description, "'", "'")
	description = strings.ReplaceAll(description, "–", "-")
	description = strings.ReplaceAll(description, "…", "...")

	return strings.TrimSpace(description)
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

package rss

import (
	"github.com/LywwKkA-aD/gocointelegraphrssparser/internal/models"
	"github.com/mmcdole/gofeed"
)

type Parser struct {
	parser *gofeed.Parser
}

func NewParser() *Parser {
	return &Parser{
		parser: gofeed.NewParser(),
	}
}

func (p *Parser) ParseFeed(url string) ([]models.NewsItem, error) {
	feed, err := p.parser.ParseURL(url)
	if err != nil {
		return nil, err
	}

	var items []models.NewsItem
	return items, nil
}

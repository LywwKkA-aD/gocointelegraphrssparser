// internal/models/news.go
package models

import "time"

type NewsItem struct {
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	Description string    `json:"description"`
	PubDate     time.Time `json:"pub_date"`
	Categories  []string  `json:"categories"`
}

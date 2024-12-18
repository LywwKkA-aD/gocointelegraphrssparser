package models

import "time"

type NewsItem struct {
    ID          string
    Title       string
    Link        string
    Description string
    PubDate     time.Time
    Categories  []string
}

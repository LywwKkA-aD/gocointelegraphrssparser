package config

import (
    "github.com/joho/godotenv"
    "os"
)

type Config struct {
    TelegramToken string
    RSSFeedURL    string
    UpdateInterval int
}

func Load() (*Config, error) {
    if err := godotenv.Load(); err != nil {
        return nil, err
    }

    return &Config{
        TelegramToken:  os.Getenv("TELEGRAM_TOKEN"),
        RSSFeedURL:     os.Getenv("RSS_FEED_URL"),
        UpdateInterval: 300,
    }, nil
}

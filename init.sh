#!/bin/bash

# Create main directory structure
mkdir -p cmd/bot \
    internal/{bot,config,models,repository,service/rss} \
    pkg/logger

# Create .gitkeep files for empty directories
find . -type d -empty -exec touch {}/.gitkeep \;

# Create main application files
cat > cmd/bot/main.go << 'EOF'
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"

    "github.com/LywwKkA-aD/gocointelegraphrssparser/internal/bot"
    "github.com/LywwKkA-aD/gocointelegraphrssparser/internal/config"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    bot, err := bot.New(cfg)
    if err != nil {
        log.Fatalf("Failed to create bot: %v", err)
    }

    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    go func() {
        sigCh := make(chan os.Signal, 1)
        signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
        <-sigCh
        cancel()
    }()

    if err := bot.Run(ctx); err != nil {
        log.Fatalf("Bot error: %v", err)
    }
}
EOF

# Create config file
cat > internal/config/config.go << 'EOF'
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
EOF

# Create models
cat > internal/models/news.go << 'EOF'
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
EOF

# Create RSS parser
cat > internal/service/rss/parser.go << 'EOF'
package rss

import (
    "github.com/mmcdole/gofeed"
    "github.com/LywwKkA-aD/gocointelegraphrssparser/internal/models"
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
EOF

# Create bot implementation
cat > internal/bot/bot.go << 'EOF'
package bot

import (
    "context"
    "time"

    tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/LywwKkA-aD/gocointelegraphrssparser/internal/config"
    "github.com/LywwKkA-aD/gocointelegraphrssparser/internal/service/rss"
)

type Bot struct {
    api    *tgbotapi.BotAPI
    parser *rss.Parser
    config *config.Config
}

func New(cfg *config.Config) (*Bot, error) {
    bot, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
    if err != nil {
        return nil, err
    }

    return &Bot{
        api:    bot,
        parser: rss.NewParser(),
        config: cfg,
    }, nil
}

func (b *Bot) Run(ctx context.Context) error {
    return nil
}
EOF

# Create env example
cat > .env.example << 'EOF'
TELEGRAM_TOKEN=your_telegram_bot_token
RSS_FEED_URL=https://cointelegraph.com/rss
EOF

# Create go.mod
cat > go.mod << 'EOF'
module github.com/LywwKkA-aD/gocointelegraphrssparser

go 1.21

require (
    github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1
    github.com/mmcdole/gofeed v1.2.1
    github.com/joho/godotenv v1.5.1
)
EOF

# Make script executable
chmod +x init.sh

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

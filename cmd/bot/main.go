// cmd/bot/main.go
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LywwKkA-aD/gocointelegraphrssparser/internal/bot"
	"github.com/LywwKkA-aD/gocointelegraphrssparser/internal/config"
	"github.com/LywwKkA-aD/gocointelegraphrssparser/pkg/logger"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load configuration: %v", err)
	}

	// Create root context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize bot
	telegramBot, err := bot.New(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize bot: %v", err)
	}

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Handle shutdown in a separate goroutine
	go handleShutdown(sigChan, cancel)

	// Log startup
	logger.Info("Starting CoinTelegraph RSS bot...")
	logger.Info("Press Ctrl+C to stop")

	// Run the bot
	if err := telegramBot.Run(ctx); err != nil {
		if err != context.Canceled {
			logger.Error("Bot stopped with error: %v", err)
			os.Exit(1)
		}
	}
}

func handleShutdown(sigChan chan os.Signal, cancel context.CancelFunc) {
	sig := <-sigChan
	logger.Info("Received signal: %v", sig)
	logger.Info("Initiating graceful shutdown...")

	// Cancel the context to initiate shutdown
	cancel()

	// Give the application some time to cleanup
	time.Sleep(3 * time.Second)

	logger.Info("Shutdown complete")
	os.Exit(0)
}

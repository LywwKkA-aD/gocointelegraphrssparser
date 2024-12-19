// internal/config/config.go
package config

import (
	"fmt"
	"os"
	"time"

	"github.com/LywwKkA-aD/gocointelegraphrssparser/pkg/logger"
	"github.com/joho/godotenv"
)

type Config struct {
	// Telegram Configuration
	TelegramToken string

	// RSS Configuration
	RSSFeedURL     string
	UpdateInterval time.Duration

	// Logger Configuration
	LogLevel    logger.Level
	LogFilePath string
}

// Load reads the configuration from environment variables
func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	cfg := &Config{}

	// Required configurations
	cfg.TelegramToken = os.Getenv("TELEGRAM_TOKEN")
	if cfg.TelegramToken == "" {
		return nil, fmt.Errorf("TELEGRAM_TOKEN is required")
	}

	cfg.RSSFeedURL = os.Getenv("RSS_FEED_URL")
	if cfg.RSSFeedURL == "" {
		cfg.RSSFeedURL = "https://cointelegraph.com/rss" // default value
	}

	// Optional configurations with defaults
	cfg.UpdateInterval = getEnvDuration("UPDATE_INTERVAL", 1*time.Minute)
	cfg.LogLevel = getEnvLogLevel("LOG_LEVEL", logger.InfoLevel)
	cfg.LogFilePath = os.Getenv("LOG_FILE")

	// Initialize logger
	if err := initLogger(cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Log configuration
	logConfig(cfg)

	return cfg, nil
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}

	duration, err := time.ParseDuration(val)
	if err != nil {
		return defaultVal
	}
	return duration
}

func getEnvLogLevel(key string, defaultVal logger.Level) logger.Level {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}

	switch val {
	case "debug":
		return logger.DebugLevel
	case "info":
		return logger.InfoLevel
	case "warn":
		return logger.WarnLevel
	case "error":
		return logger.ErrorLevel
	default:
		return defaultVal
	}
}

func initLogger(cfg *Config) error {
	return logger.Init(cfg.LogLevel, cfg.LogFilePath)
}

func logConfig(cfg *Config) {
	logger.Info("Configuration loaded:")
	logger.Info("RSS Feed URL: %s", cfg.RSSFeedURL)
	logger.Info("Update Interval: %v", cfg.UpdateInterval)
	logger.Info("Log Level: %v", cfg.LogLevel)
	if cfg.LogFilePath != "" {
		logger.Info("Log File: %s", cfg.LogFilePath)
	}
}

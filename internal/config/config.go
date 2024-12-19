package config

import (
	"fmt"
	"os"
	"time"

	"github.com/LywwKkA-aD/gocointelegraphrssparser/pkg/logger"
	"github.com/joho/godotenv"
)

type Config struct {
	TelegramToken  string
	RSSFeedURL     string
	UpdateInterval time.Duration
	LogLevel       logger.Level
	LogFilePath    string
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	cfg := &Config{}

	cfg.TelegramToken = os.Getenv("TELEGRAM_TOKEN")
	if cfg.TelegramToken == "" {
		return nil, fmt.Errorf("TELEGRAM_TOKEN is required")
	}

	cfg.RSSFeedURL = os.Getenv("RSS_FEED_URL")
	if cfg.RSSFeedURL == "" {
		cfg.RSSFeedURL = "https://cointelegraph.com/rss"
	}

	// Changed default interval to 5 seconds
	cfg.UpdateInterval = getEnvDuration("UPDATE_INTERVAL", 5*time.Second)
	cfg.LogLevel = getEnvLogLevel("LOG_LEVEL", logger.InfoLevel)
	cfg.LogFilePath = os.Getenv("LOG_FILE")

	if err := initLogger(cfg); err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

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

// internal/bot/bot.go
package bot

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LywwKkA-aD/gocointelegraphrssparser/internal/config"
	"github.com/LywwKkA-aD/gocointelegraphrssparser/internal/models"
	"github.com/LywwKkA-aD/gocointelegraphrssparser/internal/repository"
	"github.com/LywwKkA-aD/gocointelegraphrssparser/internal/service/rss"
	"github.com/LywwKkA-aD/gocointelegraphrssparser/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api      *tgbotapi.BotAPI
	parser   *rss.Parser
	cfg      *config.Config
	userRepo *repository.UserRepository
	done     chan struct{}
}

func New(cfg *config.Config) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	userRepo, err := repository.NewUserRepository("data")
	if err != nil {
		return nil, fmt.Errorf("create user repository: %w", err)
	}

	logger.Info("Authorized on account %s", api.Self.UserName)

	return &Bot{
		api:      api,
		parser:   rss.NewParser(),
		cfg:      cfg,
		userRepo: userRepo,
		done:     make(chan struct{}),
	}, nil
}

func (b *Bot) Run(ctx context.Context) error {
	logger.Info("Starting bot...")

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := b.api.GetUpdatesChan(updateConfig)
	logger.Info("Bot is ready to receive messages")

	// Start news fetching in background
	go b.fetchNewsRoutine(ctx)

	// Handle messages
	for {
		select {
		case <-ctx.Done():
			logger.Info("Shutting down bot...")
			close(b.done)
			return ctx.Err()

		case update := <-updates:
			if update.Message == nil {
				continue
			}
			go b.handleUpdate(update)
		}
	}
}

func (b *Bot) handleUpdate(update tgbotapi.Update) {
	if !update.Message.IsCommand() {
		return
	}

	logger.Debug("Received command: %s from user %d", update.Message.Command(), update.Message.Chat.ID)

	var msg tgbotapi.MessageConfig

	switch update.Message.Command() {
	case "start":
		msg = b.handleStart(update.Message)
	case "stop":
		msg = b.handleStop(update.Message)
	case "help":
		msg = b.handleHelp(update.Message)
	case "status":
		msg = b.handleStatus(update.Message)
	default:
		msg = tgbotapi.NewMessage(update.Message.Chat.ID,
			"Unknown command. Use /help to see available commands.")
	}

	if _, err := b.api.Send(msg); err != nil {
		logger.Error("Failed to send message to chat %d: %v", update.Message.Chat.ID, err)
	}
}

func (b *Bot) handleStart(message *tgbotapi.Message) tgbotapi.MessageConfig {
	users := b.userRepo.GetAll()
	if users[message.Chat.ID] {
		logger.Info("User %d attempted to subscribe again", message.Chat.ID)
		return tgbotapi.NewMessage(message.Chat.ID,
			"You are already subscribed to news updates!\n"+
				"Use /help to see available commands.")
	}

	if err := b.userRepo.Add(message.Chat.ID); err != nil {
		logger.Error("Failed to save user subscription: %v", err)
		return tgbotapi.NewMessage(message.Chat.ID,
			"Sorry, failed to subscribe. Please try again later.")
	}

	logger.Info("New user subscribed: %d", message.Chat.ID)
	return tgbotapi.NewMessage(message.Chat.ID,
		"Welcome to CoinTelegraph News Bot! 🎉\n\n"+
			"You'll now receive the latest crypto news updates.\n"+
			"Use /help to see available commands.")
}

func (b *Bot) handleStop(message *tgbotapi.Message) tgbotapi.MessageConfig {
	users := b.userRepo.GetAll()
	if !users[message.Chat.ID] {
		logger.Info("Non-subscribed user %d attempted to unsubscribe", message.Chat.ID)
		return tgbotapi.NewMessage(message.Chat.ID,
			"You are not subscribed to news updates.\n"+
				"Use /start to subscribe.")
	}

	if err := b.userRepo.Remove(message.Chat.ID); err != nil {
		logger.Error("Failed to remove user subscription: %v", err)
		return tgbotapi.NewMessage(message.Chat.ID,
			"Sorry, failed to unsubscribe. Please try again later.")
	}

	logger.Info("User unsubscribed: %d", message.Chat.ID)
	return tgbotapi.NewMessage(message.Chat.ID,
		"You've been unsubscribed from news updates.\n"+
			"Use /start to subscribe again.")
}

func (b *Bot) handleHelp(message *tgbotapi.Message) tgbotapi.MessageConfig {
	return tgbotapi.NewMessage(message.Chat.ID,
		"Available commands:\n"+
			"/start - Subscribe to news updates\n"+
			"/stop - Unsubscribe from news updates\n"+
			"/status - Check subscription status\n"+
			"/help - Show this help message")
}

func (b *Bot) handleStatus(message *tgbotapi.Message) tgbotapi.MessageConfig {
	users := b.userRepo.GetAll()
	isSubscribed := users[message.Chat.ID]

	status := "You are currently subscribed to news updates."
	if !isSubscribed {
		status = "You are not subscribed to news updates. Use /start to subscribe."
	}

	return tgbotapi.NewMessage(message.Chat.ID, status)
}

func (b *Bot) fetchNewsRoutine(ctx context.Context) {
	ticker := time.NewTicker(b.cfg.UpdateInterval)
	defer ticker.Stop()

	logger.Info("Starting news monitoring routine (interval: %v)", b.cfg.UpdateInterval)

	for {
		select {
		case <-ctx.Done():
			logger.Info("Stopping news monitoring routine")
			return

		case <-ticker.C:
			logger.Debug("Checking for new articles...")
			news, err := b.parser.FetchNews(ctx, b.cfg.RSSFeedURL)
			if err != nil {
				logger.Error("Failed to fetch news: %v", err)
				continue
			}

			if len(news) > 0 {
				logger.Info("Found %d new articles", len(news))
				b.broadcastNews(news)
			} else {
				logger.Debug("No new articles found")
			}
		}
	}
}

func (b *Bot) broadcastNews(news []models.NewsItem) {
	users := b.userRepo.GetAll()
	if len(users) == 0 {
		logger.Debug("No users to broadcast to")
		return
	}

	logger.Info("Broadcasting %d articles to %d users", len(news), len(users))

	for chatID := range users {
		for _, item := range news {
			msg := b.formatNewsMessage(item)
			msg.ChatID = chatID

			logger.Debug("Sending article '%s' to chat %d", item.Title, chatID)
			if _, err := b.api.Send(msg); err != nil {
				logger.Error("Failed to send news to chat %d: %v", chatID, err)
				continue
			}

			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (b *Bot) formatNewsMessage(item models.NewsItem) tgbotapi.MessageConfig {
	title := escapeMarkdownV2(item.Title)
	description := escapeMarkdownV2(item.Description)
	link := escapeMarkdownV2(item.Link)

	var hashtags string
	if len(item.Categories) > 0 {
		hashtags = "\n\n📌 " + formatCategories(item.Categories)
	}

	text := fmt.Sprintf(
		"🔥 *%s*\n\n%s\n\n🔗 [Read more](%s)%s",
		title,
		description,
		link,
		hashtags,
	)

	msg := tgbotapi.NewMessage(0, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2

	return msg
}

func escapeMarkdownV2(text string) string {
	// First, replace any backslashes with double backslashes
	text = strings.ReplaceAll(text, "\\", "\\\\")

	// List of special characters that need escaping
	specialChars := []string{
		"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|",
		"{", "}", ".", "!", ",",
	}

	// Escape each special character
	for _, char := range specialChars {
		text = strings.ReplaceAll(text, char, "\\"+char)
	}

	return text
}

func formatCategories(categories []string) string {
	var tags []string
	for _, cat := range categories {
		sanitized := sanitizeTag(cat)
		if sanitized != "" {
			// Escape the tag for MarkdownV2 and add hashtag
			escapedTag := escapeMarkdownV2(sanitized)
			tags = append(tags, "\\#"+escapedTag)
		}
	}

	// Limit to 5 hashtags
	if len(tags) > 5 {
		tags = tags[:5]
	}

	return strings.Join(tags, " ")
}

func sanitizeTag(tag string) string {
	// Remove any existing hashtag at the start
	tag = strings.TrimPrefix(tag, "#")
	tag = strings.TrimSpace(tag)

	// Replace spaces with underscores
	tag = strings.ReplaceAll(tag, " ", "_")

	// Keep only alphanumeric characters and underscores
	return strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return -1
	}, tag)
}

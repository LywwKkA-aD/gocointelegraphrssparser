// internal/bot/bot.go
package bot

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/LywwKkA-aD/gocointelegraphrssparser/internal/config"
	"github.com/LywwKkA-aD/gocointelegraphrssparser/internal/models"
	"github.com/LywwKkA-aD/gocointelegraphrssparser/internal/service/rss"
	"github.com/LywwKkA-aD/gocointelegraphrssparser/pkg/logger"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	api      *tgbotapi.BotAPI
	parser   *rss.Parser
	cfg      *config.Config
	users    map[int64]bool
	usersMux sync.RWMutex
	done     chan struct{}
}

func New(cfg *config.Config) (*Bot, error) {
	api, err := tgbotapi.NewBotAPI(cfg.TelegramToken)
	if err != nil {
		return nil, fmt.Errorf("create telegram bot: %w", err)
	}

	logger.Info("Authorized on account %s", api.Self.UserName)

	return &Bot{
		api:    api,
		parser: rss.NewParser(),
		cfg:    cfg,
		users:  make(map[int64]bool),
		done:   make(chan struct{}),
	}, nil
}

func (b *Bot) Run(ctx context.Context) error {
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 30

	updates := b.api.GetUpdatesChan(updateConfig)

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

	var msg tgbotapi.MessageConfig

	switch update.Message.Command() {
	case "start":
		msg = b.handleStart(update.Message)
	case "stop":
		msg = b.handleStop(update.Message)
	case "help":
		msg = b.handleHelp(update.Message)
	default:
		msg = tgbotapi.NewMessage(update.Message.Chat.ID,
			"Unknown command. Use /help to see available commands.")
	}

	if _, err := b.api.Send(msg); err != nil {
		logger.Error("Failed to send message: %v", err)
	}
}

func (b *Bot) handleStart(message *tgbotapi.Message) tgbotapi.MessageConfig {
	b.usersMux.Lock()
	b.users[message.Chat.ID] = true
	b.usersMux.Unlock()

	logger.Info("New user subscribed: %d", message.Chat.ID)
	return tgbotapi.NewMessage(message.Chat.ID,
		"Welcome to CoinTelegraph News Bot! 🎉\n\n"+
			"You'll now receive the latest crypto news updates.\n"+
			"Use /help to see available commands.")
}

func (b *Bot) handleStop(message *tgbotapi.Message) tgbotapi.MessageConfig {
	b.usersMux.Lock()
	delete(b.users, message.Chat.ID)
	b.usersMux.Unlock()

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
			"/help - Show this help message")
}

func (b *Bot) fetchNewsRoutine(ctx context.Context) {
	ticker := time.NewTicker(b.cfg.UpdateInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			news, err := b.parser.FetchNews(ctx, b.cfg.RSSFeedURL)
			if err != nil {
				logger.Error("Failed to fetch news: %v", err)
				continue
			}

			if len(news) > 0 {
				b.broadcastNews(news)
			}
		}
	}
}

func (b *Bot) broadcastNews(news []models.NewsItem) {
	b.usersMux.RLock()
	defer b.usersMux.RUnlock()

	for chatID := range b.users {
		for _, item := range news {
			msg := b.formatNewsMessage(item)
			msg.ChatID = chatID

			if _, err := b.api.Send(msg); err != nil {
				logger.Error("Failed to send news to chat %d: %v", chatID, err)
				continue
			}

			// Small delay to prevent flooding
			time.Sleep(100 * time.Millisecond)
		}
	}
}

func (b *Bot) formatNewsMessage(item models.NewsItem) tgbotapi.MessageConfig {
	var hashtags string
	if len(item.Categories) > 0 {
		hashtags = "\n\n📌 " + formatCategories(item.Categories)
	}

	text := fmt.Sprintf(
		"🔥 *%s*\n\n%s\n\n🔗 [Read more](%s)%s",
		escapeMarkdown(item.Title),
		escapeMarkdown(item.Description),
		item.Link,
		hashtags,
	)

	msg := tgbotapi.NewMessage(0, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2

	return msg
}

func escapeMarkdown(text string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(text)
}

func formatCategories(categories []string) string {
	var tags []string
	for _, cat := range categories {
		tags = append(tags, "#"+sanitizeTag(cat))
	}
	if len(tags) > 5 {
		tags = tags[:5] // Limit to 5 hashtags
	}
	return strings.Join(tags, " ")
}

func sanitizeTag(tag string) string {
	return strings.Map(func(r rune) rune {
		if r == ' ' {
			return '_'
		}
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '_' {
			return r
		}
		return -1
	}, tag)
}

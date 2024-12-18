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

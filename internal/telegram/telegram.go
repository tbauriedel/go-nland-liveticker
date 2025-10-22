package telegram

import (
	"fmt"
	"log/slog"

	t "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Bot struct {
	Instance *t.BotAPI
}

func NewBotInstance(botToken string, logger *slog.Logger) (Bot, error) {
	b := Bot{}

	instance, err := t.NewBotAPI(botToken)
	if err != nil {
		return b, fmt.Errorf("cant create telegram bot instance: %w", err)
	}

	b.Instance = instance

	logger.Info("Created telegram bot instance")

	return b, nil
}

func (b *Bot) SendMessage(chatID int64, message string) error {
	m := t.NewMessage(chatID, message)
	if _, err := b.Instance.Send(m); err != nil {
		return err
	}

	return nil
}

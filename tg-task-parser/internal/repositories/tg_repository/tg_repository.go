package tg_repository

import (
	"context"
	"log/slog"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TgRepository struct {
	bot *tgbotapi.BotAPI
}

func NewTgRepository(bot *tgbotapi.BotAPI) *TgRepository {
	return &TgRepository{
		bot: bot,
	}
}

func (r *TgRepository) SendMessage(ctx context.Context, tgID int64, text string) error {

	if tgID != 795836353 {
		return nil
	}
	slog.Info("Sending message", "tg_id", tgID, "text", text)
	msg := tgbotapi.NewMessage(tgID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := r.bot.Send(msg)
	if err != nil {
		slog.Error("Failed to send message", "tg_id", tgID, "error", err)
		return err
	}
	return nil
}

package tg_repository

import (
	"context"
	"log/slog"

	"github.com/PaulSonOfLars/gotgbot/v2"
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

	msg := tgbotapi.NewMessage(tgID, text)
	msg.ParseMode = tgbotapi.ModeMarkdown
	_, err := r.bot.Send(msg)
	if err != nil {
		slog.Error("Failed to send message", "tg_id", tgID, "error", err)
		return err
	}
	return nil
}

func (r *TgRepository) SendMessageWithButtons(chatID int64, text string, keyboard gotgbot.InlineKeyboardMarkup) error {
	// Конвертируем gotgbot клавиатуру в tgbotapi клавиатуру
	var inlineKeyboard [][]tgbotapi.InlineKeyboardButton

	for _, row := range keyboard.InlineKeyboard {
		var buttonRow []tgbotapi.InlineKeyboardButton
		for _, button := range row {
			tgButton := tgbotapi.NewInlineKeyboardButtonData(button.Text, button.CallbackData)
			buttonRow = append(buttonRow, tgButton)
		}
		inlineKeyboard = append(inlineKeyboard, buttonRow)
	}

	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = tgbotapi.ModeMarkdownV2
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(inlineKeyboard...)

	_, err := r.bot.Send(msg)
	if err != nil {
		slog.Error("Failed to send message with buttons", "chat_id", chatID, "error", err)
		return err
	}

	return nil
}

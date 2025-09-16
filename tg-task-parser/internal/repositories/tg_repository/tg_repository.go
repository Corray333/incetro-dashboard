package tg_repository

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"

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

// DownloadFile скачивает файл из Telegram по его fileID
func (r *TgRepository) DownloadFile(ctx context.Context, fileID string) ([]byte, error) {
	// Получаем информацию о файле
	fileConfig := tgbotapi.FileConfig{
		FileID: fileID,
	}

	file, err := r.bot.GetFile(fileConfig)
	if err != nil {
		slog.Error("Failed to get file info from Telegram", "fileID", fileID, "error", err)
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	// Получаем URL для скачивания файла
	fileURL := file.Link(r.bot.Token)

	// Создаем HTTP запрос с контекстом
	req, err := http.NewRequestWithContext(ctx, "GET", fileURL, nil)
	if err != nil {
		slog.Error("Failed to create download request", "fileURL", fileURL, "error", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Выполняем запрос
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("Failed to download file", "fileURL", fileURL, "error", err)
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	// Проверяем статус ответа
	if resp.StatusCode != http.StatusOK {
		slog.Error("Bad response status when downloading file", "status", resp.StatusCode, "fileURL", fileURL)
		return nil, fmt.Errorf("bad response status: %d", resp.StatusCode)
	}

	// Читаем данные файла
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Failed to read file data", "fileURL", fileURL, "error", err)
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

	slog.Info("File downloaded successfully", "fileID", fileID, "size", len(data))
	return data, nil
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

package temp_storage

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/corray333/tg-task-parser/internal/entities/message"
	"github.com/corray333/tg-task-parser/pkg/tg"
	"github.com/google/uuid"
)

// MessageProcessor обрабатывает входящие сообщения с таймерами
type MessageProcessor struct {
	storage *TempStorage
	bot     *gotgbot.Bot
}

// NewMessageProcessor создает новый процессор сообщений
func NewMessageProcessor(bot *gotgbot.Bot) *MessageProcessor {
	return &MessageProcessor{
		storage: NewTempStorage(),
		bot:     bot,
	}
}

// ProcessMessage обрабатывает входящее сообщение
func (mp *MessageProcessor) ProcessMessage(ctx context.Context, senderID, chatID int64, text string) error {
	// Создаем уникальный ID для сообщения
	msgID := fmt.Sprintf("%d_%d_%d", senderID, chatID, time.Now().UnixNano())

	// Создаем объект сообщения
	msg := &PendingMessage{
		ID:        msgID,
		SenderID:  senderID,
		Text:      text,
		ChatID:    chatID,
		Timestamp: time.Now(),
	}

	// Сохраняем сообщение во временное хранилище
	mp.storage.StorePendingMessage(msg)

	// Запускаем таймер на 2 секунды
	go mp.processAfterDelay(ctx, senderID, chatID, msg.Timestamp)

	return nil
}

// processAfterDelay обрабатывает сообщения после задержки в 2 секунды
func (mp *MessageProcessor) processAfterDelay(ctx context.Context, senderID, chatID int64, timestamp time.Time) {
	// Ждем 2 секунды
	time.Sleep(2 * time.Second)

	// Определяем временной диапазон (от времени сообщения до времени сообщения + 2 секунды)
	fromTime := timestamp
	toTime := timestamp.Add(2 * time.Second)

	// Извлекаем все сообщения от этого отправителя в указанном диапазоне
	messages := mp.storage.GetAndRemoveMessagesBySender(senderID, fromTime, toTime)

	// Если сообщений нет (уже обработаны другим таймером), завершаем
	if len(messages) == 0 {
		return
	}

	// Объединяем тексты сообщений
	combinedText := mp.combineMessages(messages)

	if !strings.Contains(combinedText, string(message.HashtagTask)) {
		return
	}

	// Создаем объединенное сообщение
	combinedMsg := &CombinedMessage{
		ID:           uuid.New(),
		SenderID:     senderID,
		ChatID:       chatID,
		CombinedText: combinedText,
		CreatedAt:    time.Now(),
	}

	// Сохраняем объединенное сообщение
	mp.storage.StoreCombinedMessage(combinedMsg)

	// Отправляем сообщение с кнопками
	if err := mp.sendMessageWithButtons(ctx, chatID, combinedText, combinedMsg.ID); err != nil {
		slog.Error("Failed to send message with buttons", "error", err)
	}
}

// combineMessages объединяет тексты сообщений
func (mp *MessageProcessor) combineMessages(messages []*PendingMessage) string {
	var texts []string
	for _, msg := range messages {
		texts = append(texts, msg.Text)
	}
	return strings.Join(texts, "\n")
}

// sendMessageWithButtons отправляет сообщение с кнопками "Принять" и "Отклонить"
func (mp *MessageProcessor) sendMessageWithButtons(ctx context.Context, chatID int64, text string, combinedMsgID uuid.UUID) error {
	// Создаем клавиатуру с кнопками
	keyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "✅ Принять",
					CallbackData: fmt.Sprintf("accept_%s", combinedMsgID.String()),
				},
				{
					Text:         "❌ Отклонить",
					CallbackData: fmt.Sprintf("reject_%s", combinedMsgID.String()),
				},
			},
		},
	}

	// Формируем текст сообщения
	messageText := fmt.Sprintf("📝 **Объединенное сообщение:**\n\n%s\n\n⏰ Выберите действие:", tg.EscapeMarkdownV2(text))

	// Отправляем сообщение
	_, err := mp.bot.SendMessage(chatID, messageText, &gotgbot.SendMessageOpts{
		ParseMode:   gotgbot.ParseModeMarkdownV2,
		ReplyMarkup: keyboard,
	})

	return err
}

// AcceptMessage обрабатывает принятие сообщения
func (mp *MessageProcessor) AcceptMessage(ctx context.Context, combinedMsgID uuid.UUID) (*CombinedMessage, error) {
	// Получаем объединенное сообщение
	combinedMsg, exists := mp.storage.GetCombinedMessage(combinedMsgID)
	if !exists {
		return nil, fmt.Errorf("combined message not found")
	}

	// Удаляем сообщение из хранилища
	mp.storage.RemoveCombinedMessage(combinedMsgID)

	return combinedMsg, nil
}

// RejectMessage обрабатывает отклонение сообщения
func (mp *MessageProcessor) RejectMessage(ctx context.Context, combinedMsgID uuid.UUID) error {
	// Просто удаляем сообщение из хранилища
	mp.storage.RemoveCombinedMessage(combinedMsgID)
	return nil
}

// GetStorage возвращает хранилище (для отладки)
func (mp *MessageProcessor) GetStorage() *TempStorage {
	return mp.storage
}

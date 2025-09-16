package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/corray333/tg-task-parser/internal/entities/message"
	"github.com/corray333/tg-task-parser/internal/repositories/temp_storage"
	"github.com/corray333/tg-task-parser/pkg/tg"
	"github.com/google/uuid"
)

// ProcessTgMessage обрабатывает сырое сообщение из Telegram
func (s *Service) ProcessTgMessage(ctx context.Context, msg *gotgbot.Message) error {
	if msg == nil {
		return nil
	}

	// Определяем ID отправителя
	senderID := msg.From.Id

	// Извлекаем текст сообщения
	text := msg.Text

	// Если есть подпись к фото, используем ее как текст
	if text == "" && msg.Caption != "" {
		text = msg.Caption
	}

	// Извлекаем информацию о изображениях
	var images []temp_storage.ImageInfo

	// Проверяем наличие фото
	if len(msg.Photo) > 0 {
		// Берем изображение с наибольшим размером (последнее в массиве)
		largestPhoto := msg.Photo[len(msg.Photo)-1]
		images = append(images, temp_storage.ImageInfo{
			FileID:       largestPhoto.FileId,
			FileUniqueID: largestPhoto.FileUniqueId,
			Width:        int(largestPhoto.Width),
			Height:       int(largestPhoto.Height),
		})
	}

	// Если есть ответ на сообщение, добавляем его к тексту и изображения
	if msg.ReplyToMessage != nil {
		replyText := msg.ReplyToMessage.Text
		if replyText == "" && msg.ReplyToMessage.Caption != "" {
			replyText = msg.ReplyToMessage.Caption
		}
		text += "\n\n" + strings.ReplaceAll(replyText, "#"+string(message.HashtagTask), "")

		// Добавляем изображения из replied сообщения
		if len(msg.ReplyToMessage.Photo) > 0 {
			largestReplyPhoto := msg.ReplyToMessage.Photo[len(msg.ReplyToMessage.Photo)-1]
			images = append(images, temp_storage.ImageInfo{
				FileID:       largestReplyPhoto.FileId,
				FileUniqueID: largestReplyPhoto.FileUniqueId,
				Width:        int(largestReplyPhoto.Width),
				Height:       int(largestReplyPhoto.Height),
			})
		}
	}

	// Обрабатываем сообщение через сервисный слой
	return s.processMessage(ctx, senderID, msg.Chat.Id, text, images)
}

// processMessage обрабатывает входящее сообщение согласно правильной архитектуре
func (s *Service) processMessage(ctx context.Context, senderID, chatID int64, text string, images []temp_storage.ImageInfo) error {
	// Создаем уникальный ID для сообщения
	msgID := fmt.Sprintf("%d_%d_%d", senderID, chatID, time.Now().UnixNano())

	// Создаем объект сообщения
	msg := &temp_storage.PendingMessage{
		ID:        msgID,
		SenderID:  senderID,
		Text:      text,
		ChatID:    chatID,
		Timestamp: time.Now(),
		Images:    images,
	}

	slog.Info("Received message", "message", msg)

	// Сохраняем сообщение во временное хранилище через репозиторий
	s.tempStorageRepo.StorePendingMessage(msg)

	// Запускаем таймер на 2 секунды через сервисный слой
	go s.processAfterDelay(ctx, senderID, chatID, msg.Timestamp)

	return nil
}

// processAfterDelay обрабатывает сообщения после задержки в 2 секунды
func (s *Service) processAfterDelay(ctx context.Context, senderID, chatID int64, timestamp time.Time) {
	// Ждем 2 секунды
	time.Sleep(2 * time.Second)

	// Определяем временной диапазон (от времени сообщения до времени сообщения + 2 секунды)
	fromTime := timestamp
	toTime := timestamp.Add(2 * time.Second)

	// Получаем все сообщения от этого отправителя в указанном диапазоне через репозиторий
	messages := s.tempStorageRepo.GetMessagesBySender(senderID, fromTime, toTime)

	slog.Info("Processing messages", "messages", messages)

	// Если сообщений нет (уже обработаны другим таймером), завершаем
	if len(messages) == 0 {
		return
	}

	// Удаляем сообщения из хранилища через репозиторий
	s.tempStorageRepo.RemoveMessagesBySender(senderID, fromTime, toTime)

	// Объединяем тексты сообщений и собираем изображения
	combinedText, allImages := s.combineMessagesWithImages(messages)

	if !strings.Contains(combinedText, "#"+string(message.HashtagTask)) {
		return
	}

	// Создаем объединенное сообщение
	combinedMsg := &temp_storage.CombinedMessage{
		ID:           uuid.New(),
		SenderID:     senderID,
		ChatID:       chatID,
		CombinedText: combinedText,
		Images:       allImages,
		CreatedAt:    time.Now(),
	}

	// Сохраняем объединенное сообщение через репозиторий
	s.tempStorageRepo.StoreCombinedMessage(combinedMsg)

	// Отправляем сообщение с кнопками через сервисный слой
	if err := s.sendMessageWithButtons(ctx, chatID, combinedText, combinedMsg.ID); err != nil {
		slog.Error("Failed to send message with buttons", "error", err)
	}
}

// combineMessagesWithImages объединяет тексты сообщений и собирает все изображения
func (s *Service) combineMessagesWithImages(messages []*temp_storage.PendingMessage) (string, []temp_storage.ImageInfo) {
	var texts []string
	var allImages []temp_storage.ImageInfo

	for _, msg := range messages {
		texts = append(texts, msg.Text)
		allImages = append(allImages, msg.Images...)
	}

	return strings.Join(texts, "\n"), allImages
}

// sendMessageWithButtons отправляет сообщение с кнопками "Принять" и "Отклонить"
func (s *Service) sendMessageWithButtons(ctx context.Context, chatID int64, text string, combinedMsgID uuid.UUID) error {
	// Создаем клавиатуру с кнопками
	keyboard := gotgbot.InlineKeyboardMarkup{
		InlineKeyboard: [][]gotgbot.InlineKeyboardButton{
			{
				{
					Text:         "✅ Принять",
					CallbackData: fmt.Sprintf("accept|%s", combinedMsgID.String()),
				},
				{
					Text:         "❌ Отклонить",
					CallbackData: fmt.Sprintf("reject|%s", combinedMsgID.String()),
				},
			},
		},
	}

	// Формируем текст сообщения с цитатой для каждой строки
	escapedText := tg.EscapeMarkdownV2(text)
	lines := strings.Split(escapedText, "\n")
	quotedLines := make([]string, len(lines))
	for i, line := range lines {
		quotedLines[i] = ">" + line
	}
	quotedText := strings.Join(quotedLines, "\n")
	messageText := fmt.Sprintf("📝 **Задача будет создана из следующего сообщения:**\n\n%s\n\n⏰ Выберите действие:", quotedText)

	// Отправляем сообщение через репозиторий Telegram
	return s.tgButtonSender.SendMessageWithButtons(chatID, messageText, keyboard)
}

// AcceptMessage обрабатывает принятие сообщения
func (s *Service) AcceptMessage(ctx context.Context, combinedMsgID uuid.UUID) (*temp_storage.CombinedMessage, error) {
	// Получаем объединенное сообщение через репозиторий
	combinedMsg, exists := s.tempStorageRepo.GetCombinedMessage(combinedMsgID)
	if !exists {
		return nil, fmt.Errorf("combined message not found")
	}

	// Удаляем сообщение из хранилища через репозиторий
	s.tempStorageRepo.RemoveCombinedMessage(combinedMsgID)

	return combinedMsg, nil
}

// RejectMessage обрабатывает отклонение сообщения
func (s *Service) RejectMessage(ctx context.Context, combinedMsgID uuid.UUID) error {
	// Просто удаляем сообщение из хранилища через репозиторий
	s.tempStorageRepo.RemoveCombinedMessage(combinedMsgID)
	return nil
}

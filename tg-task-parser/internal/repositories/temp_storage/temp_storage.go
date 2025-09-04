package temp_storage

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// PendingMessage представляет сообщение, ожидающее обработки
type PendingMessage struct {
	ID        string    // уникальный ID сообщения
	SenderID  int64     // ID отправителя (если переслано, то ID пересылающего)
	Text      string    // текст сообщения
	ChatID    int64     // ID чата
	Timestamp time.Time // время создания сообщения
}

// CombinedMessage представляет объединенное сообщение с кнопками
type CombinedMessage struct {
	ID           uuid.UUID // уникальный ID объединенного сообщения
	SenderID     int64     // ID отправителя
	ChatID       int64     // ID чата
	CombinedText string    // объединенный текст всех сообщений
	CreatedAt    time.Time // время создания объединенного сообщения
}

// TempStorage временное хранилище для сообщений
type TempStorage struct {
	pendingMessages  map[string]*PendingMessage     // ключ - ID сообщения
	combinedMessages map[uuid.UUID]*CombinedMessage // ключ - ID объединенного сообщения
	mutex            sync.RWMutex                   // мьютекс для безопасного доступа
}

// NewTempStorage создает новое временное хранилище
func NewTempStorage() *TempStorage {
	return &TempStorage{
		pendingMessages:  make(map[string]*PendingMessage),
		combinedMessages: make(map[uuid.UUID]*CombinedMessage),
	}
}

// StorePendingMessage сохраняет сообщение во временное хранилище
func (ts *TempStorage) StorePendingMessage(msg *PendingMessage) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	ts.pendingMessages[msg.ID] = msg
}

// GetAndRemoveMessagesBySender извлекает и удаляет все сообщения от указанного отправителя
// в указанном временном диапазоне
func (ts *TempStorage) GetAndRemoveMessagesBySender(senderID int64, fromTime, toTime time.Time) []*PendingMessage {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	var messages []*PendingMessage
	var toDelete []string

	for id, msg := range ts.pendingMessages {
		if msg.SenderID == senderID &&
			msg.Timestamp.After(fromTime.Add(-time.Nanosecond)) &&
			msg.Timestamp.Before(toTime.Add(time.Nanosecond)) {
			messages = append(messages, msg)
			toDelete = append(toDelete, id)
		}
	}

	// Удаляем найденные сообщения
	for _, id := range toDelete {
		delete(ts.pendingMessages, id)
	}

	return messages
}

// StoreCombinedMessage сохраняет объединенное сообщение
func (ts *TempStorage) StoreCombinedMessage(msg *CombinedMessage) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	ts.combinedMessages[msg.ID] = msg
}

// GetCombinedMessage получает объединенное сообщение по ID
func (ts *TempStorage) GetCombinedMessage(id uuid.UUID) (*CombinedMessage, bool) {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	msg, exists := ts.combinedMessages[id]
	return msg, exists
}

// RemoveCombinedMessage удаляет объединенное сообщение по ID
func (ts *TempStorage) RemoveCombinedMessage(id uuid.UUID) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	delete(ts.combinedMessages, id)
}

// GetPendingMessagesCount возвращает количество ожидающих сообщений (для отладки)
func (ts *TempStorage) GetPendingMessagesCount() int {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	return len(ts.pendingMessages)
}

// GetCombinedMessagesCount возвращает количество объединенных сообщений (для отладки)
func (ts *TempStorage) GetCombinedMessagesCount() int {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()

	return len(ts.combinedMessages)
}

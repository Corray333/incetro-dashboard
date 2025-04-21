package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/corray333/tg-task-parser/internal/entities/feedback"
)

func (r *Repository) ScanMessageMeta(ctx context.Context, chatID, messageID int64, meta *feedback.CallbackMeta) error {
	raw := json.RawMessage{}
	fmt.Println(messageID)
	if err := r.db.Get(&raw, "SELECT meta FROM message_meta WHERE chat_id = $1 AND message_id = $2", chatID, messageID); err != nil {
		slog.Error("Failed to get message meta", "error", err)
		return err
	}

	if err := json.Unmarshal(raw, meta); err != nil {
		slog.Error("Failed to unmarshal message meta", "error", err)
		return err
	}

	return nil
}

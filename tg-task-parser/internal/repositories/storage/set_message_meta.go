package storage

import (
	"context"
	"encoding/json"
	"log/slog"
)

func (r *PostgresRepository) SetMessageMeta(ctx context.Context, chatID, messageID int64, meta any) error {
	tx, isNew, err := r.getTx(ctx)
	if err != nil {
		return err
	}
	if isNew {
		defer tx.Rollback()
	}

	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	if _, err := tx.Exec("INSERT INTO message_meta (chat_id, message_id, meta) VALUES ($1, $2, $3) ON CONFLICT (chat_id, message_id) DO UPDATE SET meta = $3", chatID, messageID, metaJSON); err != nil {
		slog.Error("Failed to set message meta", "error", err)
		return err
	}

	if isNew {
		if err := tx.Commit(); err != nil {
			slog.Error("Failed to commit transaction", "error", err)
			return err
		}
	}

	return nil
}

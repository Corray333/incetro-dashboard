package storage

import (
	"context"
	"fmt"
	"log/slog"
)

func (r *PostgresRepository) UpdateEmployeeTgID(ctx context.Context, username string, tgID int64) error {
	result, err := r.db.Exec("UPDATE employees SET tg_id = $1 WHERE tg_username = $2", tgID, username)
	if err != nil {
		slog.Error("Error while updating employee tg_id", "error", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		slog.Error("Error while getting rows affected", "error", err)
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("employee with username %s not found", username)
	}

	return nil
}

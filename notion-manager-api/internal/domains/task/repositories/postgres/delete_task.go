package postgres

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

// DeleteTask deletes a task from the database by its ID
func (r *TaskPostgresRepository) DeleteTask(ctx context.Context, taskID uuid.UUID) error {
	slog.Info("Deleting task", "taskID", taskID)
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.DB().ExecContext(ctx, query, taskID)
	if err != nil {
		slog.Error("Error deleting task", "error", err)
		return err
	}
	return nil
}

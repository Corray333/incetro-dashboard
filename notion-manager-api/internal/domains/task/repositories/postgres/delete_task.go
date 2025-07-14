package postgres

import (
	"context"

	"github.com/google/uuid"
)

// DeleteTask deletes a task from the database by its ID
func (r *TaskPostgresRepository) DeleteTask(ctx context.Context, taskID uuid.UUID) error {
	query := `DELETE FROM tasks WHERE id = $1`
	_, err := r.DB().ExecContext(ctx, query, taskID)
	return err
}

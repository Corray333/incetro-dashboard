package postgres

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

// DeleteTime deletes a time entry from the database by its ID
func (r *TimePostgresRepository) DeleteTime(ctx context.Context, timeID uuid.UUID) error {
	query := `DELETE FROM times WHERE time_id = $1`
	_, err := r.DB().ExecContext(ctx, query, timeID)
	if err != nil {
		slog.Error("Error deleting time", "error", err)
		return err
	}
	return nil
}

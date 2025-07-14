package postgres

import (
	"context"

	"github.com/google/uuid"
)

// DeleteTime deletes a time entry from the database by its ID
func (r *TimePostgresRepository) DeleteTime(ctx context.Context, timeID uuid.UUID) error {
	query := `DELETE FROM times WHERE id = $1`
	_, err := r.DB().ExecContext(ctx, query, timeID)
	return err
}

package postgres

import (
	"context"
	"log/slog"

	entity_time "github.com/Corray333/employee_dashboard/internal/domains/time/entities/time"
)

func (s *TimePostgresRepository) MarkTimeWriteOfAsSent(ctx context.Context, time *entity_time.TimeOutboxMsg) error {
	tx, isNew, err := s.GetTx(ctx)
	if err != nil {
		slog.Error("error getting tx: " + err.Error())
		return err
	}
	if isNew {
		defer tx.Rollback()
	}

	if _, err := tx.Exec("DELETE FROM time_outbox WHERE time_id = $1", time.ID); err != nil {
		slog.Error("error marking time as sent: " + err.Error())
		return err
	}

	if isNew {
		if err := tx.Commit(); err != nil {
			slog.Error("error committing tx: " + err.Error())
			return err
		}
	}

	return nil
}

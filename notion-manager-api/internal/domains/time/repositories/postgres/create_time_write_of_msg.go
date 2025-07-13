package postgres

import (
	"context"
	"log/slog"

	entity_time "github.com/Corray333/employee_dashboard/internal/domains/time/entities/time"
)

func (s *TimePostgresRepository) CreateTimeWriteOf(ctx context.Context, time *entity_time.TimeOutboxMsg) error {
	tx, isNew, err := s.GetTx(ctx)
	if err != nil {
		slog.Error("error getting tx", "error", err)
		return err
	}
	if isNew {
		defer tx.Rollback()
	}

	_, err = tx.Exec("INSERT INTO time_outbox (task_id, employee_id, duration, description, work_date) VALUES ($1, $2, $3, $4, $5)", time.TaskID, time.EmployeeID, time.Duration, time.Description, time.WorkDate)
	if err != nil {
		slog.Error("Error saving time outbox message", "error", err)
		return err
	}

	if isNew {
		if err := tx.Commit(); err != nil {
			slog.Error("error committing tx", "error", err)
			return err
		}
	}

	return nil
}

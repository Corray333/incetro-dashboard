package postgres

import (
	"context"
	"log/slog"

	"github.com/Corray333/employee_dashboard/internal/domains/weekday/entities/weekday"
)

func (r *WeekdayPostgresRepository) SetWeekday(ctx context.Context, weekday *weekday.Weekday) error {
	tx, isNew, err := r.GetTx(ctx)
	if err != nil {
		return err
	}
	if isNew {
		defer tx.Rollback()
	}

	_, err = r.DB().NamedExecContext(ctx, `
		INSERT INTO weekdays (weekday_id, employee_id, category, start_time, end_time, reason, created_at, updated_at, notified)
		VALUES (:weekday_id, :employee_id, :category, :start_time, :end_time, :reason, :created_at, :updated_at, :notified)
		ON CONFLICT (weekday_id) DO UPDATE SET
			employee_id = EXCLUDED.employee_id,
			category = EXCLUDED.category,
			start_time = EXCLUDED.start_time,
			end_time = EXCLUDED.end_time,
			reason = EXCLUDED.reason,
			created_at = EXCLUDED.created_at,
			updated_at = EXCLUDED.updated_at,
			notified = EXCLUDED.notified
	`, weekdayDBFromEntity(weekday))
	if err != nil {
		slog.Error("Error setting weekday", "error", err)
		return err
	}
	if isNew {
		err = tx.Commit()
		if err != nil {
			slog.Error("Error committing transaction", "error", err)
			return err
		}
	}
	return nil
}

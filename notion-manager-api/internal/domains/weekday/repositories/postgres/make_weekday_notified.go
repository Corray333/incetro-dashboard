package postgres

import (
	"context"
	"log/slog"

	"github.com/Corray333/employee_dashboard/internal/domains/weekday/entities/weekday"
)

func (r *WeekdayPostgresRepository) MakeWeekdayNotified(ctx context.Context, weekday *weekday.Weekday) error {
	tx, isNew, err := r.GetTx(ctx)
	if err != nil {
		return err
	}
	if isNew {
		defer tx.Rollback()
	}

	if _, err := tx.ExecContext(ctx, `UPDATE weekdays SET notified = true WHERE weekday_id = $1`, weekday.ID); err != nil {
		slog.Error("Error updating weekday", "error", err)
		return err
	}

	if err != nil {
		slog.Error("Error committing transaction", "error", err)
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

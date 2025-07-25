package postgres

import (
	"context"
	"log/slog"
	_ "time"

	"github.com/Corray333/employee_dashboard/internal/domains/weekday/entities/weekday"
	_ "github.com/google/uuid"
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
		INSERT INTO weekdays (
			weekday_id, employee_id, category,
			start_time, end_time, reason,
			created_at, updated_at, notified
		)
		VALUES (
			:weekday_id, :employee_id, :category,
			:start_time, :end_time, :reason,
			:created_at, :updated_at, :notified
		)
		ON CONFLICT (weekday_id) DO UPDATE
		SET
			employee_id = EXCLUDED.employee_id,
			category    = EXCLUDED.category,
			start_time  = EXCLUDED.start_time,
			end_time    = EXCLUDED.end_time,
			reason      = EXCLUDED.reason,
			created_at  = EXCLUDED.created_at,
			updated_at  = EXCLUDED.updated_at,
			notified    = CASE
							WHEN EXCLUDED.updated_at IS DISTINCT FROM weekdays.updated_at
							THEN EXCLUDED.notified     -- время изменилось → обновляем флаг
							ELSE weekdays.notified     -- время то же → оставляем как было
						END;
	`, *weekdayDBFromEntity(weekday))
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

package postgres

import (
	"context"
	"log/slog"
	"time"

	entity_time "github.com/Corray333/employee_dashboard/internal/domains/time/entities/time"
	"github.com/google/uuid"
)

type timeWriteOfMsgDB struct {
	ID          int64     `db:"time_id"`
	TaskID      uuid.UUID `db:"task_id"`
	EmployeeID  uuid.UUID `db:"employee_id"`
	Duration    float64   `db:"duration"`
	Description string    `db:"description"`
	WorkDate    time.Time `db:"work_date"`
}

func (t *timeWriteOfMsgDB) toEntity() entity_time.TimeOutboxMsg {
	return entity_time.TimeOutboxMsg{
		ID:          t.ID,
		TaskID:      t.TaskID,
		EmployeeID:  t.EmployeeID,
		Duration:    t.Duration,
		Description: t.Description,
		WorkDate:    t.WorkDate,
	}
}

func (s *TimePostgresRepository) GetTimesMsg(ctx context.Context) (times []entity_time.TimeOutboxMsg, err error) {
	timesDB := []timeWriteOfMsgDB{}
	if err = s.DB().Select(&timesDB, "SELECT * FROM time_outbox"); err != nil {
		slog.Error("error getting times messages", "error", err)
		return nil, err
	}
	times = make([]entity_time.TimeOutboxMsg, len(timesDB))
	for i, timeDB := range timesDB {
		times[i] = timeDB.toEntity()
	}

	return times, nil
}

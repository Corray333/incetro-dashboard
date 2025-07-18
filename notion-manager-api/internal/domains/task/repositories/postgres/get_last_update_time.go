package postgres

import (
	"context"
	"log/slog"
	"time"
)

func (s *TaskPostgresRepository) GetTasksLastUpdateTime(ctx context.Context) (time.Time, error) {
	lastUpdateTime := time.Time{}
	if err := s.DB().Get(&lastUpdateTime, "SELECT COALESCE(MAX(last_edited_time), '2000-01-01 00:00:00+00') FROM tasks LIMIT 1"); err != nil {
		slog.Error("Error getting time last update time", "error", err)
		return lastUpdateTime, err
	}

	return lastUpdateTime, nil
}

func (s *TaskPostgresRepository) SetTasksLastUpdateTime(ctx context.Context, lastUpdateTime time.Time) error {
	// _, err := s.DB().Exec("UPDATE system SET tasks_db_last_sync = $1", lastUpdateTime)
	// if err != nil {
	// 	slog.Error("Error setting time last update time", "error", err)
	// 	return err
	// }
	return nil
}

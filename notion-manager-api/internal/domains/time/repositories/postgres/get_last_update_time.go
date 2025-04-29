package postgres

import (
	"context"
	"log/slog"
	"time"
)

func (s *TimePostgresRepository) GetTimeLastUpdateTime(ctx context.Context) (time.Time, error) {
	lastUpdateTime := time.Time{}
	if err := s.DB().Get(&lastUpdateTime, "SELECT times_db_last_sync FROM system LIMIT 1"); err != nil {
		slog.Error("Error getting time last update time", "error", err)
		return lastUpdateTime, err
	}

	return lastUpdateTime, nil
}

func (s *TimePostgresRepository) SetTimeLastUpdateTime(ctx context.Context, lastUpdateTime time.Time) error {
	_, err := s.DB().Exec("UPDATE system SET times_db_last_sync = $1", lastUpdateTime)
	if err != nil {
		slog.Error("Error setting time last update time", "error", err)
		return err
	}
	return nil
}

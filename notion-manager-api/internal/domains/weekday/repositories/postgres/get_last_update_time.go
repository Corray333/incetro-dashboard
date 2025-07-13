package postgres

import (
	"context"
	"log/slog"
	"time"
)

func (s *WeekdayPostgresRepository) GetWeekdaysLastUpdateTime(ctx context.Context) (time.Time, error) {
	lastUpdateTime := time.Time{}
	if err := s.DB().Get(&lastUpdateTime, "SELECT COALESCE(MAX(updated_at), '1970-01-01') FROM weekdays;"); err != nil {
		slog.Error("error getting system info", "error", err)
		return lastUpdateTime, err
	}

	return lastUpdateTime, nil
}

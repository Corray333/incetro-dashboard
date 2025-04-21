package postgres

import (
	"context"
	"log/slog"
	"time"
)

func (s *FeedbackPostgresRepository) GetFeedbackLastUpdateTime(ctx context.Context) (time.Time, error) {
	lastUpdateTime := time.Time{}
	if err := s.DB().Get(&lastUpdateTime, "SELECT feedback_db_last_sync FROM system LIMIT 1"); err != nil {
		slog.Error("error getting system info: " + err.Error())
		return lastUpdateTime, err
	}

	return lastUpdateTime, nil
}

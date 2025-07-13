package postgres

import (
	"context"
	"log/slog"
	"time"
)

func (s *ProjectPostgresRepository) GetProjectLastUpdateTime(ctx context.Context) (time.Time, error) {
	lastUpdateTime := time.Time{}
	if err := s.DB().Get(&lastUpdateTime, "SELECT project_db_last_sync FROM system LIMIT 1"); err != nil {
		slog.Error("error getting system info", "error", err)
		return lastUpdateTime, err
	}

	return lastUpdateTime, nil
}

func (s *ProjectPostgresRepository) SetProjectLastUpdateTime(ctx context.Context, lastUpdateTime time.Time) error {
	_, err := s.DB().Exec("UPDATE system SET project_db_last_sync = $1", lastUpdateTime)
	return err
}

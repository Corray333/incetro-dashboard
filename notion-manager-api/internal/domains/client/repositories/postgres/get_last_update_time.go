package postgres

import (
	"context"
	"time"
)

func (r *ClientPostgresRepository) GetClientsLastUpdateTime(ctx context.Context) (time.Time, error) {
	var lastUpdate time.Time
	query := `SELECT clients_db_last_sync FROM system LIMIT 1`
	err := r.DB().GetContext(ctx, &lastUpdate, query)
	if err != nil {
		return time.Time{}, err
	}
	return lastUpdate, nil
}

package postgres

import (
	"github.com/Corray333/employee_dashboard/internal/postgres"
)

type TimePostgresRepository struct {
	*postgres.PostgresClient
}

func NewTimePostgresRepository(client *postgres.PostgresClient) *TimePostgresRepository {
	return &TimePostgresRepository{client}
}

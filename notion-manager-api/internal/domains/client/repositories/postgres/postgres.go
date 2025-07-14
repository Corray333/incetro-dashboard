package postgres

import (
	"github.com/Corray333/employee_dashboard/internal/postgres"
)

type ClientPostgresRepository struct {
	*postgres.PostgresClient
}

func NewClientPostgresRepository(client *postgres.PostgresClient) *ClientPostgresRepository {
	return &ClientPostgresRepository{
		PostgresClient: client,
	}
}
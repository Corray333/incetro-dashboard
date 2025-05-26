package postgres

import (
	"github.com/Corray333/employee_dashboard/internal/postgres"
)

type ProjectPostgresRepository struct {
	*postgres.PostgresClient
}

func NewProjectPostgresRepository(client *postgres.PostgresClient) *ProjectPostgresRepository {
	return &ProjectPostgresRepository{client}
}

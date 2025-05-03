package postgres

import (
	"github.com/Corray333/employee_dashboard/internal/postgres"
)

type TaskPostgresRepository struct {
	*postgres.PostgresClient
}

func NewTaskPostgresRepository(client *postgres.PostgresClient) *TaskPostgresRepository {
	return &TaskPostgresRepository{client}
}

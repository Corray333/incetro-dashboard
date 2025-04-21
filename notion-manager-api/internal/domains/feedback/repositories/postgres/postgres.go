package postgres

import (
	"github.com/Corray333/employee_dashboard/internal/postgres"
)

type FeedbackPostgresRepository struct {
	*postgres.PostgresClient
}

func NewFeedbackPostgresRepository(client *postgres.PostgresClient) *FeedbackPostgresRepository {
	return &FeedbackPostgresRepository{client}
}

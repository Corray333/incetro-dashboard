package notion

import (
	notion "github.com/Corray333/employee_dashboard/pkg/notion/v2"
)

type TimeNotionRepository struct {
	client *notion.Client
}

func NewTimeNotionRepository(client *notion.Client) *TimeNotionRepository {
	return &TimeNotionRepository{
		client: client,
	}
}

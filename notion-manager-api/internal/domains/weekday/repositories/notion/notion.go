package notion

import (
	notion "github.com/Corray333/employee_dashboard/pkg/notion/v2"
)

type WeekdayNotionRepository struct {
	client *notion.Client
}

func NewWeekdayNotionRepository(client *notion.Client) *WeekdayNotionRepository {
	return &WeekdayNotionRepository{
		client: client,
	}
}

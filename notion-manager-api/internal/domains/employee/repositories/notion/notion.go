package notion

import (
	notion "github.com/Corray333/employee_dashboard/pkg/notion/v2"
)

type EmployeeNotionRepository struct {
	client *notion.Client
}

func NewWeekdayNotionRepository(client *notion.Client) *EmployeeNotionRepository {
	return &EmployeeNotionRepository{
		client: client,
	}
}

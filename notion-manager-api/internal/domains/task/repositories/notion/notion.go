package notion

import (
	notion "github.com/Corray333/employee_dashboard/pkg/notion/v2"
)

type TaskNotionRepository struct {
	client *notion.Client
}

func NewTaskNotionRepository(client *notion.Client) *TaskNotionRepository {
	return &TaskNotionRepository{
		client: client,
	}
}

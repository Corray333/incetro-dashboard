package notion

import (
	notion "github.com/Corray333/employee_dashboard/pkg/notion/v2"
)

type ProjectNotionRepository struct {
	client *notion.Client
}

func NewProjectNotionRepository(client *notion.Client) *ProjectNotionRepository {
	return &ProjectNotionRepository{
		client: client,
	}
}

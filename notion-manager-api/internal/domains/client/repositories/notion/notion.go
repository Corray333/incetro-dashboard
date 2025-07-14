package notion

import (
	notion "github.com/Corray333/employee_dashboard/pkg/notion/v2"
)

type ClientNotionRepository struct {
	client *notion.Client
}

func NewClientNotionRepository(client *notion.Client) *ClientNotionRepository {
	return &ClientNotionRepository{
		client: client,
	}
}
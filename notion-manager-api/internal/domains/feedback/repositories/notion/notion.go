package notion

import (
	notion "github.com/Corray333/employee_dashboard/pkg/notion/v2"
)

type FeedbackNotionRepository struct {
	client *notion.Client
}

func NewFeedbackNotionRepository(client *notion.Client) *FeedbackNotionRepository {
	return &FeedbackNotionRepository{
		client: client,
	}
}

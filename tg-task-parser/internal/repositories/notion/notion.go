package notion

import (
	"context"
	"strings"

	notion "github.com/corray333/tg-task-parser/pkg/notion/v2"
	"github.com/google/uuid"
)

type NotionRepository struct {
	client *notion.Client
}

func NewNotionRepository(client *notion.Client) *NotionRepository {
	return &NotionRepository{
		client: client,
	}
}

func (r *NotionRepository) AnswerFeedback(ctx context.Context, feedbackID uuid.UUID, answer string) error {
	content := []map[string]interface{}{
		{
			"object": "block",
			"type":   "paragraph",
			"paragraph": map[string]interface{}{
				"rich_text": []map[string]interface{}{
					{
						"type": "text",
						"text": map[string]string{
							"content": answer,
						},
					},
				},
			},
		},
	}

	_, err := r.client.AppendPageContent(strings.ReplaceAll(feedbackID.String(), "-", ""), content)
	if err != nil {
		return err
	}

	return nil
}

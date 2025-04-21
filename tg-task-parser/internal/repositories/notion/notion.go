package notion

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	notion "github.com/corray333/tg-task-parser/pkg/notion/v2"
	"github.com/google/uuid"
	"github.com/spf13/viper"
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

	_, err := r.client.AddCommentToPage(strings.ReplaceAll(feedbackID.String(), "-", ""), answer)
	if err != nil {
		return err
	}

	return nil
}

func (r *NotionRepository) NewFeedback(ctx context.Context, feedback string) (uuid.UUID, error) {

	newPageProperties := map[string]interface{}{
		"Name": map[string]interface{}{
			"title": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]string{
						"content": feedback,
					},
				},
			},
		},
	}

	page := struct {
		ID string `json:"id"`
	}{}
	resp, err := r.client.CreatePage(viper.GetString("notion.databases.feedback"), newPageProperties, nil)
	if err != nil {
		return uuid.Nil, err
	}

	if err := json.Unmarshal(resp, &page); err != nil {
		slog.Error("Failed to unmarshal page", "error", err)
		return uuid.Nil, err
	}

	pageID, err := uuid.Parse(page.ID)
	if err != nil {
		slog.Error("Failed to parse page ID", "error", err)
		return uuid.Nil, err
	}

	return pageID, nil

}

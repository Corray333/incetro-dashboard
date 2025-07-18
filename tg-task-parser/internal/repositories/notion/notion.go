package notion

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/corray333/tg-task-parser/internal/entities/topic"
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

func (r *NotionRepository) NewFeedback(ctx context.Context, projectID uuid.UUID, feedback string) (uuid.UUID, error) {

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
		"Проект": map[string]interface{}{
			"relation": []map[string]interface{}{
				{
					"id": projectID.String(),
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

type topicNotion struct {
	Properties struct {
		Name struct {
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"Name"`
		Icon struct {
			RichText []struct {
				PlainText string `json:"plain_text"`
			} `json:"rich_text"`
		} `json:"Icon"`
	} `json:"properties"`
}

func (t *topicNotion) toEntity() *topic.Topic {
	icon := ""
	for _, r := range t.Properties.Icon.RichText {
		icon += r.PlainText
	}
	return &topic.Topic{
		Name: t.Properties.Name.Title[0].PlainText,
		Icon: icon,
	}
}

func (r *NotionRepository) GetTopics(ctx context.Context) ([]topic.Topic, error) {

	resp, err := r.client.SearchPages(viper.GetString("notion.databases.topics"), nil)
	if err != nil {
		return nil, err
	}
	topicsRaw := []topicNotion{}

	if err := json.Unmarshal(resp, &topicsRaw); err != nil {
		slog.Error("Error unmarshalling feedbacks from notion", "error", err)
		return nil, err
	}

	topics := []topic.Topic{}
	for _, f := range topicsRaw {
		topics = append(topics, *f.toEntity())
	}

	return topics, nil
}

func (r *NotionRepository) GetEmployeesWithIncorrectTimeEntries(ctx context.Context) ([]uuid.UUID, error) {
	req := map[string]interface{}{
		"filter": map[string]interface{}{
			"property": "Ошибки",
			"formula": map[string]interface{}{
				"string": map[string]interface{}{
					"is_not_empty": true,
				},
			},
		},
	}
	resp, err := r.client.SearchPages(viper.GetString("notion.databases.times"), req)
	if err != nil {
		slog.Error("Error searching incorrect times", "error", err)
		return nil, err
	}

	// Парсим JSON ответ
	var response []struct {
		Properties struct {
			Person struct {
				Relation []struct {
					ID string `json:"id"`
				} `json:"relation"`
			} `json:"Person"`
		} `json:"properties"`
	}

	if err := json.Unmarshal(resp, &response); err != nil {
		slog.Error("Error unmarshaling response", "error", err)
		return nil, err
	}

	// Извлекаем уникальные UUID сотрудников
	employeeMap := make(map[uuid.UUID]bool)
	for _, result := range response {
		for _, person := range result.Properties.Person.Relation {
			employeeUUID, err := uuid.Parse(person.ID)
			if err != nil {
				slog.Error("Error parsing employee UUID", "error", err, "id", person.ID)
				continue
			}
			employeeMap[employeeUUID] = true
		}
	}

	// Преобразуем map в slice
	employees := make([]uuid.UUID, 0, len(employeeMap))
	for employeeUUID := range employeeMap {
		employees = append(employees, employeeUUID)
	}

	return employees, nil
}

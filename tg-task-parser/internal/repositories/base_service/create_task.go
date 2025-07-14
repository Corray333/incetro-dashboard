package base_service

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/corray333/tg-task-parser/internal/entities/message"
	"github.com/corray333/tg-task-parser/pkg/notion"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type PageCreated struct {
	ID string `json:"id"`
}

func (r *BaseService) CreateTask(ctx context.Context, task *message.Message, projectID uuid.UUID) (string, error) {
	req := map[string]interface{}{
		"Task": map[string]interface{}{
			"type": "title",
			"title": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]interface{}{
						"content": task.Text,
					},
				},
			},
		},
		"Исполнитель": map[string]interface{}{
			"type": "people",
			"people": func() []map[string]interface{} {
				executor := []map[string]interface{}{}
				for _, user := range task.Mentions {
					employeeID, err := r.GetEmployeeByTgUsername(ctx, string(user))
					if err != nil {
						slog.Error("Notion error while getting employee ID: " + err.Error())
						return nil
					}
					executor = append(executor, map[string]interface{}{
						"id": employeeID,
					})
				}
				return executor
			}(),
		},
		"Теги": map[string]interface{}{
			"type": "multi_select",
			"multi_select": func() []map[string]interface{} {
				tags := []map[string]interface{}{}
				for _, tag := range task.Hashtags {
					tags = append(tags, map[string]interface{}{
						"name": tag,
					})
				}
				return tags
			}(),
		},
		"Продукт": map[string]interface{}{
			"type": "relation",
			"relation": []map[string]interface{}{
				{
					"id": projectID.String(),
				},
			},
		},
	}

	// Создаем страницу задачи в Notion
	res, err := notion.CreatePage(viper.GetString("notion.databases.tasks"), req, nil, "")
	if err != nil {
		slog.Error("Notion error while creating task: " + err.Error())
		return "", err
	}

	var pageCreated PageCreated
	if err := json.Unmarshal(res, &pageCreated); err != nil {
		slog.Error("Error while unmarshaling page created: " + err.Error())
		return "", err
	}

	return pageCreated.ID, nil
}

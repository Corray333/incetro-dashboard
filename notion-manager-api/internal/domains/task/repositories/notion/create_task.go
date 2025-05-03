package notion

import (
	"context"
	"log/slog"
	"time"

	entity_task "github.com/Corray333/employee_dashboard/internal/domains/task/entities/task"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

func (r *TaskNotionRepository) CreateTask(ctx context.Context, task *entity_task.Task) error {
	req := map[string]interface{}{
		"Task": map[string]interface{}{
			"type": "title",
			"title": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]interface{}{
						"content": task.Task,
					},
				},
			},
		},
		"Статус": map[string]interface{}{
			"type": "status",
			"status": map[string]interface{}{
				"name": entity_task.StatusCanDo,
			},
		},
		"Оценка": map[string]interface{}{
			"type":   "number",
			"number": task.Estimate,
		},
		"Теги": map[string]interface{}{
			"type": "multi_select",
			"multi_select": func() []map[string]interface{} {
				res := make([]map[string]interface{}, 0, len(task.Tags))
				for _, tag := range task.Tags {
					res = append(res, map[string]interface{}{
						"name": tag,
					})
				}
				return res
			}(),
		},
		"Дедлайн": map[string]interface{}{
			"type": "date",
			"date": map[string]interface{}{
				"start": task.Start.Format(time.RFC3339),
				"end":   task.End.Format(time.RFC3339),
			},
		},
		"Исполнитель": map[string]interface{}{
			"type": "people",
			"people": func() []map[string]interface{} {
				res := make([]map[string]interface{}, 0, len(task.ExecutorIDs))
				for _, executorID := range task.ExecutorIDs {
					res = append(res, map[string]interface{}{
						"id": executorID.String(),
					})
				}
				return res
			}(),
		},
		"Ответственный": map[string]interface{}{
			"type": "people",
			"people": func() []map[string]interface{} {
				res := make([]map[string]interface{}, 0, len(task.ExecutorIDs))
				for _, executorID := range task.ExecutorIDs {
					res = append(res, map[string]interface{}{
						"id": executorID.String(),
					})
				}
				return res
			}(),
		},
		"Продукт": map[string]interface{}{
			"type": "relation",
			"relation": []map[string]interface{}{
				{
					"id": task.ProjectID.String(),
				},
			},
		},
		"Приоритет": map[string]interface{}{
			"type": "select",
			"select": map[string]interface{}{
				"name": task.Priority,
			},
		},
	}

	if _, err := r.client.CreatePage(viper.GetString("notion.databases.tasks"), req, nil); err != nil {
		slog.Error("Error writing time of", "error", err)
		return err
	}

	return nil
}

type taskNotion struct {
	ID             string `json:"id"`
	CreatedTime    string `json:"created_time"`
	LastEditedTime string `json:"last_edited_time"`
	Properties     struct {
		Status struct {
			Status struct {
				Name string `json:"name"`
			} `json:"status"`
		} `json:"Статус"`

		Time struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Время"`

		ParentTask struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Родительская задача"`

		Previous struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Предыдущая"`

		Responsible struct {
			People []struct {
				ID string `json:"id"`
			} `json:"people"`
		} `json:"Ответственный"`

		Creator struct {
			ID string `json:"id"`
		} `json:"Кто создал"`

		Executor struct {
			People []struct {
				ID string `json:"id"`
			} `json:"people"`
		} `json:"Исполнитель"`

		// Questions struct {
		// 	Relation []struct {
		// 		ID string `json:"id"`
		// 	} `json:"relation"`
		// } `json:"Вопросы"`

		Product struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Продукт"`

		Estimate struct {
			Number float64 `json:"number"`
		} `json:"Оценка"`

		Tags struct {
			MultiSelect []struct {
				Name string `json:"name"`
			} `json:"multi_select"`
		} `json:"Теги"`

		Deadline struct {
			Date struct {
				Start string `json:"start"`
				End   string `json:"end"`
			} `json:"date"`
		} `json:"Дедлайн"`

		Task struct {
			Title []textElement `json:"title"`
		} `json:"Task"`

		Priority struct {
			Select struct {
				Name string `json:"name"`
			} `json:"select"`
		} `json:"Приоритет"`
	} `json:"properties"`
}

func (t *taskNotion) toEntity() *entity_task.Task {
	entity := &entity_task.Task{
		ID:             parseUUIDOrNil(t.ID),
		CreatedTime:    t.CreatedTime,
		LastEditedTime: t.LastEditedTime,
		Priority:       t.Properties.Priority.Select.Name,
		Task:           getPlainText(t.Properties.Task.Title),
		Status:         entity_task.Status(t.Properties.Status.Status.Name),
		Estimate:       t.Properties.Estimate.Number,
		Tags:           make([]string, 0, len(t.Properties.Tags.MultiSelect)),
	}

	if len(t.Properties.ParentTask.Relation) > 0 {
		entity.ParentID = parseUUIDOrNil(t.Properties.ParentTask.Relation[0].ID)
	}

	if len(t.Properties.Responsible.People) > 0 {
		entity.Responsible = make([]uuid.UUID, 0, len(t.Properties.Responsible.People))
		for _, person := range t.Properties.Responsible.People {
			entity.Responsible = append(entity.Responsible, parseUUIDOrNil(person.ID))
		}
	}

	if len(t.Properties.Executor.People) > 0 {
		entity.ExecutorIDs = make([]uuid.UUID, 0, len(t.Properties.Executor.People))
		for _, person := range t.Properties.Executor.People {
			entity.ExecutorIDs = append(entity.ExecutorIDs, parseUUIDOrNil(person.ID))
		}
	}

	if len(t.Properties.Product.Relation) > 0 {
		entity.ProjectID = parseUUIDOrNil(t.Properties.Product.Relation[0].ID)
	}

	for _, tag := range t.Properties.Tags.MultiSelect {
		entity.Tags = append(entity.Tags, tag.Name)
	}

	if t.Properties.Deadline.Date.Start != "" {
		if deadlineStart, err := time.Parse(time.RFC3339, t.Properties.Deadline.Date.Start); err == nil {
			entity.Start = deadlineStart
		}
	}

	if t.Properties.Deadline.Date.End != "" {
		if deadlineEnd, err := time.Parse(time.RFC3339, t.Properties.Deadline.Date.End); err == nil {
			entity.End = deadlineEnd
		}
	}

	return entity
}

// parseUUIDOrNil безопасно парсит UUID, возвращая uuid.Nil при ошибке
func parseUUIDOrNil(s string) uuid.UUID {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil
	}
	return id
}

type textElement struct {
	PlainText string `json:"plain_text"`
}

// getFirstPlainText возвращает первый PlainText, если он есть, иначе пустую строку
func getPlainText(titles []textElement) string {
	res := ""
	for _, title := range titles {
		res = title.PlainText
	}
	return res
}

package notion

import (
	"context"
	"log/slog"
	"time"

	entity_task "github.com/Corray333/employee_dashboard/internal/domains/task/entities/task"
	"github.com/google/uuid"
	"github.com/nav-inc/datetime"
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
				return []map[string]interface{}{
					{
						"id": task.ExecutorID.String(),
					},
				}
			}(),
		},
		"Ответственный": map[string]interface{}{
			"type": "people",
			"people": func() []map[string]interface{} {
				return []map[string]interface{}{
					{
						"id": task.ExecutorID.String(),
					},
				}
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

type Formula struct {
	Type    string  `json:"type"`
	Number  float64 `json:"number,omitempty"`
	String  string  `json:"string,omitempty"`
	Boolean bool    `json:"boolean,omitempty"`
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

		TotalHours struct {
			Formula Formula `json:"formula"`
		} `json:"Тотал ч."`

		Subtasks struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Подзадачи"`

		Next struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Следующая"`

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

		TBH struct {
			Formula Formula `json:"formula"`
		} `json:"TBH"`

		MainTask struct {
			Formula Formula `json:"formula"`
		} `json:"Главная задача"`

		CP struct {
			Formula Formula `json:"formula"`
		} `json:"CP"`

		Priority struct {
			Select struct {
				Name string `json:"name"`
			} `json:"select"`
		} `json:"Приоритет"`

		InProgress struct {
			Formula Formula `json:"formula"`
		} `json:"В работе"`

		TotalEstimate struct {
			Formula Formula `json:"formula"`
		} `json:"Тотал оценка"`

		PlanFact struct {
			Formula Formula `json:"formula"`
		} `json:"План / Факт"`

		Section struct {
			Formula Formula `json:"formula"`
		} `json:"Секция"`

		Duration struct {
			Formula Formula `json:"formula"`
		} `json:"Длительность"`

		Chain struct {
			Formula Formula `json:"formula"`
		} `json:"Цепочка проектов"`

		Volume struct {
			Formula Formula `json:"formula"`
		} `json:"Объем"`

		CR struct {
			Formula Formula `json:"formula"`
		} `json:"CR"`

		Questions struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Вопросы"`

		IKP struct {
			Select struct {
				Name string `json:"name"`
			} `json:"select"`
		} `json:"IKP"`

		CreatedDate struct {
			CreatedTime string `json:"created_time"`
		} `json:"Дата создания"`

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
	} `json:"properties"`
}

func (t *taskNotion) toEntity() *entity_task.Task {
	createdTime, err := datetime.Parse(t.CreatedTime, time.UTC)
	if err != nil {
		slog.Error("Error parsing created time", "error", err)
		return nil
	}

	lastEditedTime, err := datetime.Parse(t.LastEditedTime, time.UTC)
	if err != nil {
		slog.Error("Error parsing last edited time", "error", err)
		return nil
	}

	entity := &entity_task.Task{
		ID:             parseUUIDOrNil(t.ID),
		CreatedTime:    createdTime,
		LastEditedTime: lastEditedTime,
		Priority:       t.Properties.Priority.Select.Name,
		Task:           getPlainText(t.Properties.Task.Title),
		Status:         entity_task.Status(t.Properties.Status.Status.Name),
		Estimate:       t.Properties.Estimate.Number,
		Tags:           make([]entity_task.Tag, 0, len(t.Properties.Tags.MultiSelect)),
		CreatorID:      parseUUIDOrNil(t.Properties.Creator.ID),
		Start:          time.Time{},
		End:            time.Time{},
		TotalHours:     t.Properties.TotalHours.Formula.Number,
		TBH:            t.Properties.TBH.Formula.Number,
		CP:             t.Properties.CP.Formula.Number,
		TotalEstimate:  t.Properties.TotalEstimate.Formula.Number,
		PlanFact:       t.Properties.PlanFact.Formula.Number,
		Duration:       t.Properties.Duration.Formula.Number,
		CR:             t.Properties.CR.Formula.Number,
		IKP:            t.Properties.IKP.Select.Name,
		MainTask:       t.Properties.MainTask.Formula.String,
	}

	if len(t.Properties.ParentTask.Relation) > 0 {
		entity.ParentID = parseUUIDOrNil(t.Properties.ParentTask.Relation[0].ID)
	}

	if len(t.Properties.Responsible.People) > 0 {
		entity.ResponsibleID = parseUUIDOrNil(t.Properties.Responsible.People[0].ID)
	}

	if len(t.Properties.Executor.People) > 0 {
		entity.ExecutorID = parseUUIDOrNil(t.Properties.Executor.People[0].ID)
	}

	if len(t.Properties.Product.Relation) > 0 {
		entity.ProjectID = parseUUIDOrNil(t.Properties.Product.Relation[0].ID)
	}

	for _, tag := range t.Properties.Tags.MultiSelect {
		entity.Tags = append(entity.Tags, entity_task.Tag(tag.Name))
	}

	if t.Properties.Deadline.Date.Start != "" {
		if deadlineStart, err := datetime.Parse(t.Properties.Deadline.Date.Start, time.UTC); err == nil {
			entity.Start = deadlineStart
		}
	}

	if t.Properties.Deadline.Date.End != "" {
		if deadlineEnd, err := datetime.Parse(t.Properties.Deadline.Date.End, time.UTC); err == nil {
			entity.End = deadlineEnd
		}
	}

	if len(t.Properties.Subtasks.Relation) > 0 {
		entity.Subtasks = make([]uuid.UUID, 0, len(t.Properties.Subtasks.Relation))
		for _, subtask := range t.Properties.Subtasks.Relation {
			entity.Subtasks = append(entity.Subtasks, parseUUIDOrNil(subtask.ID))
		}
	}

	if len(t.Properties.Previous.Relation) > 0 {
		entity.PreviousID = parseUUIDOrNil(t.Properties.Previous.Relation[0].ID)
	}

	if len(t.Properties.Next.Relation) > 0 {
		entity.NextID = parseUUIDOrNil(t.Properties.Next.Relation[0].ID)
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

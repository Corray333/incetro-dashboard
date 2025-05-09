package notion

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	pkg_time "time"

	entity_time "github.com/Corray333/employee_dashboard/internal/domains/time/entities/time"
	"github.com/Corray333/employee_dashboard/pkg/notion"
	"github.com/google/uuid"
	"github.com/nav-inc/datetime"
	"github.com/spf13/viper"
)

func (r *TimeNotionRepository) ListTimes(ctx context.Context, lastUpdate pkg_time.Time) ([]entity_time.Time, error) {
	filter := map[string]interface{}{
		"filter": map[string]interface{}{
			"timestamp": "last_edited_time",
			"last_edited_time": map[string]interface{}{
				"after": lastUpdate.Format(notion.TIME_LAYOUT),
			},
		},
		"sorts": []map[string]interface{}{
			{
				"timestamp": "created_time",
				"direction": "ascending",
			},
		},
	}

	resp, err := r.client.SearchPages(viper.GetString("notion.databases.times"), filter)
	if err != nil {
		slog.Error("Error getting times from notion", "error", err)
		return nil, err
	}
	feedbacksRaw := []time{}

	if err := json.Unmarshal(resp, &feedbacksRaw); err != nil {
		slog.Error("Error unmarshalling feedbacks from notion", "error", err)
		return nil, err
	}

	feedbacks := []entity_time.Time{}
	for _, f := range feedbacksRaw {
		feedbacks = append(feedbacks, *f.ToEntity())
	}

	return feedbacks, nil

}

type time struct {
	ID             string `json:"id"`
	CreatedTime    string `json:"created_time"`
	LastEditedTime string `json:"last_edited_time"`
	Properties     struct {
		TotalHours struct {
			Number float64 `json:"number"`
		} `json:"Затрачено ч."`
		Analytics struct {
			Relation []struct{} `json:"relation"`
		} `json:"Аналитика"`
		PayableHours struct {
			Formula struct {
				Number float64 `json:"number"`
			} `json:"formula"`
		} `json:"К оплате ч."`
		Task struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Задача"`
		Direction struct {
			Select struct {
				Name string `json:"name"`
			} `json:"select"`
		} `json:"Направление"`
		TaskName struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Название задачи"`
		WorkDate struct {
			Date struct {
				Start    string      `json:"start"`
				End      interface{} `json:"end"`
				TimeZone interface{} `json:"time_zone"`
			} `json:"date"`
		} `json:"Дата работ"`
		WhoDid struct {
			People []struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			} `json:"people"`
		} `json:"Исполнитель"`
		CreatedTimeField struct {
			CreatedTime string `json:"created_time"`
		} `json:"Created time"`
		Payment struct {
			Checkbox bool `json:"checkbox"`
		} `json:"Оплата"`
		Project struct {
			Rollup struct {
				Array []struct {
					Relation []struct {
						ID string `json:"id"`
					} `json:"relation"`
				} `json:"array"`
			} `json:"rollup"`
		} `json:"Проект"`
		StatusHours struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Статус ч"`
		Month struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Месяц"`
		ProjectName struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Имя проекта"`
		ProjectStatus struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Статус проекта"`
		WhatDid struct {
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"Что делали"`
		BH struct {
			Formula struct {
				Number float64 `json:"number"`
			} `json:"formula"`
		} `json:"BH"`
		SH struct {
			Number float64 `json:"number"` // Number or null
		} `json:"SH"`
		DH struct {
			Number float64 `json:"number"` // Number or null
		} `json:"DH"`
		BHGS struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"BHGS"`
		WeekNumber struct {
			Formula struct {
				Number float64 `json:"number"`
			} `json:"formula"`
		} `json:"Номер недели"`
		DayNumber struct {
			Formula struct {
				Number float64 `json:"number"`
			} `json:"formula"`
		} `json:"Номер дня"`
		MonthNumber struct {
			Formula struct {
				Number float64 `json:"number"`
			} `json:"formula"`
		} `json:"Номер месяца"`
		// Добавленные поля:
		PH struct {
			Formula struct {
				Number float64 `json:"number"`
			} `json:"formula"`
		} `json:"PH"`
		Expertise struct {
			Rollup struct {
				Array []struct {
					Relation []struct {
						ID string `json:"id"`
					} `json:"relation"`
				} `json:"array"`
			} `json:"rollup"`
		} `json:"Экспертиза"`
		Overtime struct {
			Checkbox bool `json:"checkbox"`
		} `json:"Сверхурочные"`
		PCB struct {
			Formula struct {
				Boolean bool `json:"boolean"`
			} `json:"formula"`
		} `json:"PC-B"`
		TaskEstimate struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Оценка задачи"`
		Person struct {
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
		} `json:"Person"`
		IDField struct {
			UniqueID struct {
				Prefix string `json:"prefix"`
				Number int    `json:"number"`
			} `json:"unique_id"`
		} `json:"ID"`
		ET struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"ET"`
		Priority struct {
			Rollup struct {
				Array []struct {
					Select struct {
						Name string `json:"name"`
					} `json:"select"`
				} `json:"array"`
			} `json:"rollup"`
		} `json:"Приоритет"`
		MainTask struct {
			Rollup struct {
				Array []struct {
					Formula struct {
						String string `json:"string"`
					} `json:"formula"`
				} `json:"array"`
			} `json:"rollup"`
		} `json:"Главная задача"`
		TargetTask struct {
			Formula struct {
				String string `json:"string"`
			} `json:"formula"`
		} `json:"Целевая задача"`
		CR struct {
			Formula struct {
				Boolean bool `json:"boolean"`
			} `json:"formula"`
		} `json:"CR"`
	} `json:"properties"`
	URL string `json:"url"`
}

func (t *time) ToEntity() *entity_time.Time {
	lastUpdate, err := datetime.Parse(t.LastEditedTime, pkg_time.UTC)
	if err != nil {
		lastUpdate = pkg_time.Time{}
	}
	created, err := datetime.Parse(t.CreatedTime, pkg_time.UTC)
	if err != nil {
		created = pkg_time.Time{}
	}
	workDate, err := datetime.Parse(t.Properties.WorkDate.Date.Start, pkg_time.UTC)
	if err != nil {
		workDate = pkg_time.Time{}
	}
	var taskID uuid.UUID
	if len(t.Properties.Task.Relation) > 0 {
		tID, err := uuid.Parse(t.Properties.Task.Relation[0].ID)
		if err != nil {
			taskID = uuid.Nil
		}
		taskID = tID
	}
	var projectID uuid.UUID
	if len(t.Properties.Project.Rollup.Array) > 0 && len(t.Properties.Project.Rollup.Array[0].Relation) > 0 {
		pID, err := uuid.Parse(t.Properties.Project.Rollup.Array[0].Relation[0].ID)
		if err != nil {
			projectID = uuid.Nil
		}
		projectID = pID
	}
	var employeeID uuid.UUID
	if len(t.Properties.WhoDid.People) > 0 {
		eID, err := uuid.Parse(t.Properties.WhoDid.People[0].ID)
		if err != nil {
			employeeID = uuid.Nil
		}
		employeeID = eID
	}
	var expertiseID uuid.UUID
	if len(t.Properties.Expertise.Rollup.Array) > 0 && len(t.Properties.Expertise.Rollup.Array[0].Relation) > 0 {
		exID, err := uuid.Parse(t.Properties.Expertise.Rollup.Array[0].Relation[0].ID)
		if err != nil {
			expertiseID = uuid.Nil
		}
		expertiseID = exID
	}
	var personID uuid.UUID
	if len(t.Properties.Person.Relation) > 0 {
		pID, err := uuid.Parse(t.Properties.Person.Relation[0].ID)
		if err != nil {
			personID = uuid.Nil
		}
		personID = pID
	}
	id, _ := uuid.Parse(t.ID)
	return &entity_time.Time{
		ID:            id,
		TotalHours:    t.Properties.TotalHours.Number,
		PayableHours:  t.Properties.PayableHours.Formula.Number,
		TaskID:        taskID,
		Direction:     t.Properties.Direction.Select.Name,
		WorkDate:      workDate,
		EmployeeID:    employeeID,
		Payment:       t.Properties.Payment.Checkbox,
		ProjectID:     projectID,
		StatusHours:   t.Properties.StatusHours.Formula.String,
		Month:         t.Properties.Month.Formula.String,
		ProjectName:   t.Properties.ProjectName.Formula.String,
		ProjectStatus: t.Properties.ProjectStatus.Formula.String,
		WhatDid: func() string {
			if len(t.Properties.WhatDid.Title) > 0 {
				return t.Properties.WhatDid.Title[0].PlainText
			}
			return ""
		}(),
		BH:          t.Properties.BH.Formula.Number,
		SH:          t.Properties.SH.Number,
		DH:          t.Properties.DH.Number,
		BHGS:        t.Properties.BHGS.Formula.String,
		WeekNumber:  t.Properties.WeekNumber.Formula.Number,
		DayNumber:   t.Properties.DayNumber.Formula.Number,
		MonthNumber: t.Properties.MonthNumber.Formula.Number,
		PH:          t.Properties.PH.Formula.Number,
		ExpertiseID: expertiseID,
		Overtime:    t.Properties.Overtime.Checkbox,
		PCB:         t.Properties.PCB.Formula.Boolean,
		PersonID:    personID,
		IDField:     t.Properties.IDField.UniqueID.Prefix + "-" + fmt.Sprintf("%d", t.Properties.IDField.UniqueID.Number),
		ET:          t.Properties.ET.Formula.String,
		Priority: func() string {
			if len(t.Properties.Priority.Rollup.Array) > 0 && len(t.Properties.Priority.Rollup.Array[0].Select.Name) > 0 {
				return t.Properties.Priority.Rollup.Array[0].Select.Name
			}
			return ""
		}(),
		MainTask: func() string {
			if len(t.Properties.MainTask.Rollup.Array) > 0 && len(t.Properties.MainTask.Rollup.Array[0].Formula.String) > 0 {
				return t.Properties.MainTask.Rollup.Array[0].Formula.String
			}
			return ""
		}(),
		TargetTask: t.Properties.TargetTask.Formula.String,
		CR:         t.Properties.CR.Formula.Boolean,
		LastUpdate: lastUpdate,
		CreatedAt:  created,
	}
}

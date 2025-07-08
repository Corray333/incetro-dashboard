package notion

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/employee/entities/employee"
	"github.com/Corray333/employee_dashboard/internal/domains/weekday/entities/weekday"
	"github.com/Corray333/employee_dashboard/pkg/notion"
	"github.com/google/uuid"
	"github.com/nav-inc/datetime"
	"github.com/spf13/viper"
)

func (r *WeekdayNotionRepository) ListWeekdays(ctx context.Context, lastUpdate time.Time) ([]weekday.Weekday, error) {
	filter := map[string]interface{}{
		"filter": map[string]interface{}{
			"timestamp": "last_edited_time",
			"on_or_after": map[string]interface{}{
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

	resp, err := notion.SearchPages(viper.GetString("notion.databases.weekday"), filter)
	if err != nil {
		return nil, err
	}
	weekdaysRaw := struct {
		Results []Weekday `json:"results"`
	}{}

	if err := json.Unmarshal(resp, &weekdaysRaw); err != nil {
		slog.Error("Error unmarshalling weekdays from notion", "error", err)
		return nil, err
	}

	weekday := []weekday.Weekday{}
	for _, f := range weekdaysRaw.Results {
		weekday = append(weekday, *f.ToEntity())
	}

	return weekday, nil

}

type Weekday struct {
	ID             uuid.UUID `json:"id"`
	CreatedTime    string    `json:"created_time"`
	LastEditedTime string    `json:"last_edited_time"`
	Properties     struct {
		Employee struct {
			Relation []struct {
				ID uuid.UUID `json:"id"`
			} `json:"relation"`
		} `json:"Сотрудник"`
		TotalDays struct {
			Formula struct {
				Number int `json:"number"`
			} `json:"formula"`
		} `json:"Всего дней"`
		Category struct {
			Select struct {
				Name string `json:"name"`
			} `json:"select"`
		} `json:"Категория"`
		Period struct {
			Date struct {
				Start    string `json:"start"`
				End      string `json:"end"`
				TimeZone string `json:"time_zone"`
			} `json:"date"`
		} `json:"Период"`
		Reason struct {
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"Причина"`
	} `json:"properties"`
}

func (w *Weekday) ToEntity() *weekday.Weekday {
	createdTime, err := datetime.Parse(w.CreatedTime, time.UTC)
	if err != nil {
		slog.Error("Error parsing created time", "error", err)
		return nil
	}
	lastEditedTime, err := datetime.Parse(w.LastEditedTime, time.UTC)
	if err != nil {
		slog.Error("Error parsing last edited time", "error", err)
		return nil
	}

	periodStart, err := datetime.Parse(w.Properties.Period.Date.Start, time.UTC)
	if err != nil {
		slog.Error("Error parsing period start time", "error", err)
		return nil
	}

	periodEnd := time.Time{}
	if w.Properties.Period.Date.End != "" {
		periodEnd, err = datetime.Parse(w.Properties.Period.Date.End, time.UTC)
		if err != nil {
			slog.Error("Error parsing period end time", "error", err)
			return nil
		}
	}

	employeeID := uuid.Nil
	if len(w.Properties.Employee.Relation) > 0 {
		employeeID = w.Properties.Employee.Relation[0].ID
	}

	reason := ""
	if len(w.Properties.Reason.Title) > 0 {
		reason = w.Properties.Reason.Title[0].PlainText
	}

	return &weekday.Weekday{
		ID:          w.ID,
		Category:    weekday.Category(w.Properties.Category.Select.Name),
		PeriodStart: periodStart,
		PeriodEnd:   periodEnd,
		Reason:      reason,
		CreatedAt:   createdTime,
		UpdatedAt:   lastEditedTime,
		Employee: employee.Employee{
			ID: employeeID,
		},
	}
}

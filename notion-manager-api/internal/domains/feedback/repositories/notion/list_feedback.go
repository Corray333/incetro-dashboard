package notion

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/feedback/entities/feedback"
	"github.com/Corray333/employee_dashboard/pkg/notion"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

func (r *FeedbackNotionRepository) ListFeedback(ctx context.Context, lastUpdate time.Time) ([]feedback.Feedback, error) {
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

	resp, err := notion.SearchPages(viper.GetString("notion.databases.feedback"), filter)
	if err != nil {
		return nil, err
	}
	feedbacksRaw := struct {
		Results []Feedback `json:"results"`
	}{}

	if err := json.Unmarshal(resp, &feedbacksRaw); err != nil {
		slog.Error("Error unmarshalling feedbacks from notion", slog.String("error", err.Error()))
		return nil, err
	}

	feedbacks := []feedback.Feedback{}
	for _, f := range feedbacksRaw.Results {
		feedbacks = append(feedbacks, *f.ToEntity())
	}

	return feedbacks, nil

}

type Feedback struct {
	Object         string `json:"object"`
	ID             string `json:"id"`
	CreatedTime    string `json:"created_time"`
	LastEditedTime string `json:"last_edited_time"`
	CreatedBy      struct {
		Object string `json:"object"`
		ID     string `json:"id"`
	} `json:"created_by"`
	LastEditedBy struct {
		Object string `json:"object"`
		ID     string `json:"id"`
	} `json:"last_edited_by"`
	Cover *interface{} `json:"cover"`
	Icon  struct {
		Type     string `json:"type"`
		External struct {
			URL string `json:"url"`
		} `json:"external"`
	} `json:"icon"`
	Parent struct {
		Type       string `json:"type"`
		DatabaseID string `json:"database_id"`
	} `json:"parent"`
	Archived   bool `json:"archived"`
	InTrash    bool `json:"in_trash"`
	Properties struct {
		Type struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Select struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"select"`
		} `json:"Тип"`
		Priority struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Select struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"select"`
		} `json:"Приоритет"`
		Task struct {
			ID       string        `json:"id"`
			Type     string        `json:"type"`
			Relation []interface{} `json:"relation"`
			HasMore  bool          `json:"has_more"`
		} `json:"Задача"`
		Project struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
			HasMore bool `json:"has_more"`
		} `json:"Проект"`
		CreatedDate struct {
			ID          string `json:"id"`
			Type        string `json:"type"`
			CreatedTime string `json:"created_time"`
		} `json:"Дата создания"`
		Direction struct {
			ID          string        `json:"id"`
			Type        string        `json:"type"`
			MultiSelect []interface{} `json:"multi_select"`
		} `json:"Направление"`
		Status struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Status struct {
				ID    string `json:"id"`
				Name  string `json:"name"`
				Color string `json:"color"`
			} `json:"status"`
		} `json:"Статус"`
		Name struct {
			ID    string `json:"id"`
			Type  string `json:"type"`
			Title []struct {
				Type string `json:"type"`
				Text struct {
					Content string      `json:"content"`
					Link    interface{} `json:"link"`
				} `json:"text"`
				Annotations struct {
					Bold          bool   `json:"bold"`
					Italic        bool   `json:"italic"`
					Strikethrough bool   `json:"strikethrough"`
					Underline     bool   `json:"underline"`
					Code          bool   `json:"code"`
					Color         string `json:"color"`
				} `json:"annotations"`
				PlainText string      `json:"plain_text"`
				Href      interface{} `json:"href"`
			} `json:"title"`
		} `json:"Name"`
	} `json:"properties"`
	URL       string      `json:"url"`
	PublicURL interface{} `json:"public_url"`
}

func (f *Feedback) ToEntity() *feedback.Feedback {
	return &feedback.Feedback{
		Text: func() string {
			if len(f.Properties.Name.Title) == 0 {
				return ""
			}
			return f.Properties.Name.Title[0].PlainText
		}(),
		Type: func() string {
			if f.Properties.Type.Select.Name == "" {
				return ""
			}
			return f.Properties.Type.Select.Name
		}(),
		Priority: func() string {
			if f.Properties.Priority.Select.Name == "" {
				return ""
			}
			return f.Properties.Priority.Select.Name
		}(),
		ProjectID: func() uuid.UUID {
			if len(f.Properties.Project.Relation) == 0 {
				return uuid.UUID{}
			}
			id, err := uuid.Parse(f.Properties.Project.Relation[0].ID)
			if err != nil {
				return uuid.Nil
			}
			return id
		}(),
		TaskID: func() uuid.UUID {
			if len(f.Properties.Task.Relation) == 0 {
				return uuid.UUID{}
			}
			return uuid.UUID{}
		}(),
		CreatedDate: func() time.Time {
			createdDate, err := time.Parse(notion.TIME_LAYOUT_IN, f.Properties.CreatedDate.CreatedTime)
			if err != nil {
				return time.Time{}
			}
			return createdDate
		}(),
		Direction: func() string {
			if len(f.Properties.Direction.MultiSelect) == 0 {
				return ""
			}
			return ""
		}(),
		Status: f.Properties.Status.Status.Name,
		ID: func() uuid.UUID {
			id, err := uuid.Parse(f.ID)
			if err != nil {
				return uuid.Nil
			}
			return id
		}(),
	}
}

package notion

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/project/entities/project"
	"github.com/Corray333/employee_dashboard/internal/utils"
	"github.com/Corray333/employee_dashboard/pkg/notion"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

func (r *ProjectNotionRepository) ListFeedback(ctx context.Context, lastUpdate time.Time) ([]project.Project, error) {
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

	resp, err := notion.SearchPages(viper.GetString("notion.databases.project"), filter)
	if err != nil {
		return nil, err
	}
	feedbacksRaw := []projectNotion{}

	if err := json.Unmarshal(resp, &feedbacksRaw); err != nil {
		slog.Error("Error unmarshalling feedbacks from notion", "error", err)
		return nil, err
	}

	feedbacks := []project.Project{}
	for _, f := range feedbacksRaw {
		feedbacks = append(feedbacks, *f.ToEntity())
	}

	return feedbacks, nil

}

type projectNotion struct {
	ID             string `json:"id"`
	CreatedTime    string `json:"created_time"`
	LastEditedTime string `json:"last_edited_time"`
	Icon           struct {
		Type     string `json:"type"`
		External struct {
			Url string `json:"url"`
		} `json:"external"`
		File struct {
			Url string `json:"url"`
		} `json:"file"`
		Emoji string `json:"emoji"`
	} `json:"icon"`
	Properties struct {
		Name struct {
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"Name"`
		Status struct {
			Status struct {
				Name  string `json:"name"`
				ID    string `json:"id"`
				Color string `json:"color"`
			} `json:"status"`
		} `json:"Статус"`
		ProjectType struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Select struct {
				Name string `json:"name"`
			} `json:"select"`
		} `json:"Тип проекта"`
		Manager struct {
			ID       string `json:"id"`
			Type     string `json:"type"`
			Relation []struct {
				ID string `json:"id"`
			} `json:"relation"`
			HasMore bool `json:"has_more"`
		} `json:"Менеджер"`
		ManagerLink struct {
			ID     string `json:"id"`
			Type   string `json:"type"`
			Rollup struct {
				Type     string        `json:"type"`
				Array    []interface{} `json:"array"` // Определите структуру при необходимости
				Function string        `json:"function"`
			} `json:"rollup"`
		} `json:"Менеджер Link"`
		GSL struct {
			Type string `json:"type"`
			URL  string `json:"url"`
		} `json:"GSL"`
	} `json:"properties"`
}

func (p *projectNotion) ToEntity() *project.Project {
	lastEdited, err := time.Parse(notion.TIME_LAYOUT_IN, p.LastEditedTime)
	if err != nil {
		return nil
	}

	return &project.Project{
		ID: utils.ParseUUIDOrNil(p.ID),
		Name: func() string {
			if len(p.Properties.Name.Title) == 0 {
				return ""
			}
			return p.Properties.Name.Title[0].PlainText
		}(),
		Icon: func() string {
			if p.Icon.Type == "emoji" {
				return p.Icon.Emoji
			}
			if p.Icon.Type == "external" {
				return p.Icon.External.Url
			}
			if p.Icon.Type == "file" {
				return p.Icon.File.Url
			}
			return ""
		}(),
		IconType: p.Icon.Type,
		Status:   p.Properties.Status.Status.Name,
		ManagerID: func() uuid.UUID {
			if len(p.Properties.Manager.Relation) == 0 {
				return uuid.Nil
			}
			return utils.ParseUUIDOrNil(p.Properties.Manager.Relation[0].ID)
		}(),
		Type:       p.Properties.ProjectType.Select.Name,
		SheetsLink: p.Properties.GSL.URL,
		UpdatedAt:  lastEdited,
	}
}

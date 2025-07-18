package notion

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/client/entities/client"
	"github.com/Corray333/employee_dashboard/pkg/notion"
	"github.com/google/uuid"
	"github.com/nav-inc/datetime"
	"github.com/spf13/viper"
)

func (r *ClientNotionRepository) ListClients(ctx context.Context, lastUpdate time.Time) ([]client.Client, error) {
	filter := map[string]interface{}{
		"filter": map[string]interface{}{
			"timestamp": "last_edited_time",
			"last_edited_time": map[string]interface{}{
				"on_or_after": lastUpdate.Format(notion.TIME_LAYOUT),
			},
		},
		"sorts": []map[string]interface{}{
			{
				"timestamp": "last_edited_time",
				"direction": "ascending",
			},
		},
	}

	resp, err := notion.SearchPages(viper.GetString("notion.databases.client"), filter)
	if err != nil {
		return nil, err
	}
	clientsRaw := struct {
		Results []Client `json:"results"`
	}{}

	if err := json.Unmarshal(resp, &clientsRaw); err != nil {
		slog.Error("Error unmarshalling clients from notion", "error", err)
		return nil, err
	}

	clients := []client.Client{}
	for _, c := range clientsRaw.Results {
		clients = append(clients, *c.ToEntity())
	}

	return clients, nil
}

type Client struct {
	ID             uuid.UUID `json:"id"`
	CreatedTime    string    `json:"created_time"`
	LastEditedTime string    `json:"last_edited_time"`
	Properties     struct {
		ID struct {
			UniqueID struct {
				Prefix string `json:"prefix"`
				Number int64  `json:"number"`
			} `json:"unique_id"`
		} `json:"ID"`
		Source struct {
			Select struct {
				Name string `json:"name"`
			} `json:"select"`
		} `json:"Откуда пришел"`
		Status struct {
			Status struct {
				Name string `json:"name"`
			} `json:"status"`
		} `json:"Статус"`
		Projects struct {
			Relation []struct {
				ID uuid.UUID `json:"id"`
			} `json:"relation"`
		} `json:"Проекты"`
		Name struct {
			Title []struct {
				PlainText string `json:"plain_text"`
			} `json:"title"`
		} `json:"Клиент"`
	} `json:"properties"`
}

func (c *Client) ToEntity() *client.Client {
	createdTime, err := datetime.Parse(c.CreatedTime, time.UTC)
	if err != nil {
		slog.Error("Error parsing created time", "error", err)
		return nil
	}
	lastEditedTime, err := datetime.Parse(c.LastEditedTime, time.UTC)
	if err != nil {
		slog.Error("Error parsing last edited time", "error", err)
		return nil
	}

	name := ""
	if len(c.Properties.Name.Title) > 0 {
		name = c.Properties.Name.Title[0].PlainText
	}

	uniqueID := c.Properties.ID.UniqueID.Number

	projectIDs := make([]uuid.UUID, 0, len(c.Properties.Projects.Relation))
	for _, project := range c.Properties.Projects.Relation {
		projectIDs = append(projectIDs, project.ID)
	}

	return &client.Client{
		ID:         c.ID,
		Name:       name,
		Status:     client.Status(c.Properties.Status.Status.Name),
		Source:     c.Properties.Source.Select.Name,
		UniqueID:   uniqueID,
		CreatedAt:  createdTime,
		UpdatedAt:  lastEditedTime,
		ProjectIDs: projectIDs,
	}
}

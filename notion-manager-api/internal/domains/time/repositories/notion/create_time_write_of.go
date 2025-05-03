package notion

import (
	"context"
	"log/slog"
	"math"

	entity_time "github.com/Corray333/employee_dashboard/internal/domains/time/entities/time"
	notion "github.com/Corray333/employee_dashboard/pkg/notion/v2"
	"github.com/spf13/viper"
)

func (r *TimeNotionRepository) CreateTimeWriteOf(ctx context.Context, writeOf *entity_time.TimeOutboxMsg) error {
	req := map[string]interface{}{
		"Затрачено ч.": map[string]interface{}{
			"number": math.Ceil((float64(writeOf.Duration)/60/60)/0.25) * 0.25,
		},
		"Задача": map[string]interface{}{
			"relation": []map[string]interface{}{
				{
					"id": writeOf.TaskID,
				},
			},
		},
		"Что делали": map[string]interface{}{
			"type": "title",
			"title": []map[string]interface{}{
				{
					"type": "text",
					"text": map[string]interface{}{
						"content": writeOf.Description,
					},
				},
			},
		},
		"Дата работ": map[string]interface{}{
			"type": "date",
			"date": map[string]interface{}{
				"start": writeOf.WorkDate.Format(notion.TIME_LAYOUT),
			},
		},
		"Исполнитель": map[string]interface{}{
			"people": []map[string]interface{}{
				{
					"object": "user",
					"id":     writeOf.EmployeeID,
				},
			},
		},
	}

	if _, err := r.client.CreatePage(viper.GetString("notion.databases.times"), req, nil); err != nil {
		slog.Error("Error writing time of", "error", err)
		return err
	}

	return nil

}

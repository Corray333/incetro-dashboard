package notion

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/task/entities/task"
	notion "github.com/Corray333/employee_dashboard/pkg/notion/v2"
	"github.com/spf13/viper"
)

func (r *TaskNotionRepository) ListTasks(ctx context.Context, lastUpdate time.Time) ([]task.Task, error) {
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

	resp, err := r.client.SearchPages(viper.GetString("notion.databases.tasks"), filter)
	if err != nil {
		return nil, err
	}
	tasksRaw := []taskNotion{}

	if err := json.Unmarshal(resp, &tasksRaw); err != nil {
		slog.Error("Error unmarshalling feedbacks from notion", "error", err)
		return nil, err
	}

	tasks := []task.Task{}
	for _, f := range tasksRaw {
		tasks = append(tasks, *f.toEntity())
	}

	return tasks, nil
}

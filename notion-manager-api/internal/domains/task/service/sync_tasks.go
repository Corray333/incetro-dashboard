package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/project/entities/project"
	"github.com/Corray333/employee_dashboard/internal/domains/task/entities/task"
	"github.com/Corray333/employee_dashboard/internal/utils"
	"github.com/google/uuid"
	"github.com/spf13/viper"
)

type taskSetter interface {
	SetTask(ctx context.Context, task *task.Task) error
}

type notionTaskLister interface {
	ListTasks(ctx context.Context, lastUpdate time.Time) ([]task.Task, error)
}

// SetTasksLastUpdateTime(ctx context.Context, lastUpdateTime time.Time) error
// GetTasksLastUpdateTime(ctx context.Context) (time.Time, error)

type tasksLastUpdateGetter interface {
	GetTasksLastUpdateTime(ctx context.Context) (time.Time, error)
}

type tasksLastUpdateSetter interface {
	SetTasksLastUpdateTime(ctx context.Context, lastUpdateTime time.Time) error
}

func (s *TaskService) TaskSync(ctx context.Context) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		if err := s.updateTasks(ctx); err != nil {
			slog.Error("Notion tasks sync error", "error", err)
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (s *TaskService) updateTasks(ctx context.Context) error {
	slog.Info("Updating tasks")
	lastUpdateTime, err := s.tasksLastUpdateGetter.GetTasksLastUpdateTime(ctx)
	if err != nil {
		return err
	}

	times, err := s.notionTaskLister.ListTasks(ctx, lastUpdateTime)
	if err != nil {
		return err
	}
	if len(times) == 0 {
		slog.Info("No tasks found")
		return nil
	}
	slog.Info("Found tasks", "count", len(times))

	lastTime := time.Time{}
	for _, task := range times {
		if err := s.taskSetter.SetTask(ctx, &task); err != nil {
			return err
		}
		if task.LastEditedTime.After(lastTime) {
			lastTime = task.LastEditedTime
		}
	}
	if err := s.tasksLastUpdateSetter.SetTasksLastUpdateTime(ctx, lastTime); err != nil {
		return err
	}

	return nil
}

type taskLister interface {
	ListTasks(ctx context.Context, filter task.Filter, limit, offset int) ([]task.Task, error)
}

type sheetsTasksUpdater interface {
	UpdateSheetsTasks(ctx context.Context, sheetID string, tasks []task.Task) error
}

type projectsLister interface {
	ListProjects(ctx context.Context) ([]project.Project, error)
}

func (s *TaskService) UpdateProjectSheets(ctx context.Context, projectID uuid.UUID) error {
	tasks, err := s.taskLister.ListTasks(ctx, task.Filter{ProjectID: projectID}, 20000, 0)
	if err != nil {
		return err
	}

	projects, err := s.projectsLister.ListProjects(ctx)
	if err != nil {
		return err
	}

	var sheetID string
	for _, project := range projects {
		if project.ID == projectID {
			sheetID, err = utils.ExtractSpreadsheetID(project.SheetsLink)
			if err != nil {
				return err
			}
			break
		}
	}

	if sheetID == "" {
		return nil
	}

	if err := s.sheetsTasksUpdater.UpdateSheetsTasks(ctx, sheetID, tasks); err != nil {
		return err
	}

	return nil

}

func (s *TaskService) updateSheets(ctx context.Context) {

	tasks, err := s.taskLister.ListTasks(ctx, task.Filter{}, 20000, 0)
	if err != nil {
		slog.Error("Error getting tasks", "error", err)
		return
	}

	if err := s.sheetsTasksUpdater.UpdateSheetsTasks(ctx, viper.GetString("sheets.id"), tasks); err != nil {
		slog.Error("Error updating sheets", "error", err)
		return
	}

	// projects, err := s.projectsLister.ListProjects(ctx)
	// if err != nil {
	// 	slog.Error("Error getting projects", "error", err)
	// 	return
	// }

	// for _, project := range projects {
	// 	if project.SheetsLink == "" {
	// 		continue
	// 	}

	// 	sheetID, err := utils.ExtractSpreadsheetID(project.SheetsLink)
	// 	if err != nil {
	// 		slog.Error("Error extracting spreadsheet ID", "error", err)
	// 		continue
	// 	}

	// 	tasks, err := s.taskLister.ListTasks(ctx, task.Filter{ProjectID: project.ID}, 20000, 0)
	// 	if err != nil {
	// 		slog.Error("Error getting tasks", "error", err)
	// 		return
	// 	}

	// 	if err := s.sheetsTasksUpdater.UpdateSheetsTasks(ctx, sheetID, tasks); err != nil {
	// 		slog.Error("Error updating sheets", "error", err)
	// 		return
	// 	}

	// }
}

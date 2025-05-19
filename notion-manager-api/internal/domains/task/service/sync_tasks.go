package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/task/entities/task"
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
			slog.Error("Notion time sync error", "error", err)
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
		return nil
	}

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
	ListTasks(ctx context.Context, limit, offset int) ([]task.Task, error)
}

type sheetsTasksUpdater interface {
	UpdateSheetsTasks(ctx context.Context, sheetID string, tasks []task.Task) error
}

func (s *TaskService) updateSheets(ctx context.Context) {

	times, err := s.taskLister.ListTasks(ctx, 20000, 0)
	if err != nil {
		slog.Error("Error getting tasks", "error", err)
		return
	}

	if err := s.sheetsTasksUpdater.UpdateSheetsTasks(ctx, viper.GetString("sheets.id"), times); err != nil {
		slog.Error("Error updating sheets", "error", err)
		return
	}
}

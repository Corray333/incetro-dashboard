package service

import (
	"context"
	"log/slog"
	"time"

	entity_task "github.com/Corray333/employee_dashboard/internal/domains/task/entities/task"
)

type TaskService struct {
	taskOutboxMsgGetter  taskOutboxMsgGetter
	taskOutboxMsgDeleter taskOutboxMsgDeleter
	notionTaskCreator    notionTaskCreator
	taskMsgCreator       taskMsgCreator

	taskSetter            taskSetter
	notionTaskLister      notionTaskLister
	tasksLastUpdateGetter tasksLastUpdateGetter
	tasksLastUpdateSetter tasksLastUpdateSetter

	taskLister         taskLister
	sheetsTasksUpdater sheetsTasksUpdater
}

type postgresRepository interface {
	taskOutboxMsgGetter
	taskOutboxMsgDeleter
	taskMsgCreator
	taskSetter
	tasksLastUpdateGetter
	tasksLastUpdateSetter
	taskLister
}

type notionRepository interface {
	notionTaskCreator
	notionTaskLister
}

type sheetsRepository interface {
	sheetsTasksUpdater
}

type option func(*TaskService)

func NewTaskService(opts ...option) *TaskService {
	service := &TaskService{}

	for _, opt := range opts {
		opt(service)
	}

	service.updateSheets(context.Background())

	return service
}

func (s *TaskService) AcceptUpdate(ctx context.Context) {
	go s.updateSheets(ctx)
}

func WithPostgresRepository(repository postgresRepository) option {
	return func(s *TaskService) {
		s.taskOutboxMsgGetter = repository
		s.taskOutboxMsgDeleter = repository
		s.taskMsgCreator = repository
		s.taskSetter = repository
		s.tasksLastUpdateGetter = repository
		s.tasksLastUpdateSetter = repository
		s.taskLister = repository
	}
}

func WithNotionRepository(repository notionRepository) option {
	return func(s *TaskService) {
		s.notionTaskCreator = repository
		s.notionTaskLister = repository
	}
}

func WithSheetsRepository(repository sheetsRepository) option {
	return func(s *TaskService) {
		s.sheetsTasksUpdater = repository
	}
}

func (s *TaskService) Run() {
	go s.TaskSync(context.Background())
	go s.StartTaskOutboxWorker(context.Background())
}

type taskOutboxMsgGetter interface {
	GetTaskOutboxMsgs(ctx context.Context) ([]entity_task.TaskOutboxMsg, error)
}

type notionTaskCreator interface {
	CreateTask(ctx context.Context, task *entity_task.Task) error
}

type taskOutboxMsgDeleter interface {
	DeleteTaskOutboxMsg(ctx context.Context, task *entity_task.TaskOutboxMsg) error
}

func (s *TaskService) processTaskMsgs(ctx context.Context) error {
	taskMsgs, err := s.taskOutboxMsgGetter.GetTaskOutboxMsgs(ctx)
	if err != nil {
		return err
	}

	for _, task := range taskMsgs {
		if err := s.notionTaskCreator.CreateTask(ctx, task.ToEntity()); err != nil {
			return err
		}
		if err := s.taskOutboxMsgDeleter.DeleteTaskOutboxMsg(ctx, &task); err != nil {
			return err
		}
	}

	return nil
}

func (s *TaskService) StartTaskOutboxWorker(ctx context.Context) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		if err := s.processTaskMsgs(ctx); err != nil {
			slog.Error("Error processing new task messages", "error", err)
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return
		}
	}
}

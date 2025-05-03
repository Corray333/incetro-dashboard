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
}

type postgresRepository interface {
	taskOutboxMsgGetter
	taskOutboxMsgDeleter
	taskMsgCreator
}
type notionRepository interface {
	notionTaskCreator
}

type option func(*TaskService)

func NewTaskService(opts ...option) *TaskService {
	service := &TaskService{}

	for _, opt := range opts {
		opt(service)
	}

	return service
}

func (s *TaskService) AcceptUpdate(ctx context.Context) {

}

func WithPostgresRepository(repository postgresRepository) option {
	return func(s *TaskService) {
		s.taskOutboxMsgGetter = repository
		s.taskOutboxMsgDeleter = repository
		s.taskMsgCreator = repository
	}
}

func WithNotionRepository(repository notionRepository) option {
	return func(s *TaskService) {
		s.notionTaskCreator = repository
	}
}

func (s *TaskService) Run() {
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

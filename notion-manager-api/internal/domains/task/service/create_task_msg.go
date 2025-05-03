package service

import (
	"context"

	entity_task "github.com/Corray333/employee_dashboard/internal/domains/task/entities/task"
)

type taskMsgCreator interface {
	CreateTaskOutboxMsg(ctx context.Context, msg *entity_task.TaskOutboxMsg) error
}

func (s *TaskService) CreateTask(ctx context.Context, task *entity_task.TaskOutboxMsg) error {
	return s.taskMsgCreator.CreateTaskOutboxMsg(ctx, task)
}

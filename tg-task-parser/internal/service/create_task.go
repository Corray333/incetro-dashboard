package service

import (
	"context"
	"log/slog"

	"github.com/corray333/tg-task-parser/internal/entities/task"
)

type taskCreator interface {
	CreateTask(ctx context.Context, task *task.Task) error
}

func (s *Service) CreateTask(ctx context.Context, message string, replyMessage string) error {
	newTask, err := task.TaskFromMessage(message, replyMessage)
	if err != nil {
		slog.Error("error while creating task from message", "error", err)
		return err
	}
	if err := s.taskCreator.CreateTask(ctx, newTask); err != nil {
		slog.Error("error while creating task in repository", "error", err)
		return err
	}

	return nil
}

package service

import (
	"context"
	"log/slog"
	"slices"

	"github.com/corray333/tg-task-parser/internal/entities/message"
	"github.com/google/uuid"
)

type taskCreator interface {
	CreateTask(ctx context.Context, task *message.Message, projectID uuid.UUID) (string, error)
}

type projectByChatIDGetter interface {
	GetProjectByChatID(ctx context.Context, chatID int64) (uuid.UUID, error)
}

func (s *Service) CreateTask(ctx context.Context, chatID int64, msg string, replyMessage string) (string, error) {
	projectID, err := s.projectByChatIDGetter.GetProjectByChatID(ctx, chatID)
	if err != nil {
		return "", err
	}

	newTask, err := message.ParseMessage(msg, replyMessage)
	if err != nil {
		return "", err
	}

	if !slices.Contains(newTask.Hashtags, message.HashtagTask) {
		return "", nil
	}
	pageID, err := s.taskCreator.CreateTask(ctx, newTask, projectID)
	if err != nil {
		slog.Error("error while creating task in repository", "error", err)
		return "", err
	}

	return pageID, nil
}

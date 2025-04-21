package service

import (
	"context"

	"github.com/corray333/tg-task-parser/internal/entities/feedback"
	"github.com/corray333/tg-task-parser/internal/entities/message"
	"github.com/google/uuid"
)

type feedbackCreator interface {
	NewFeedback(ctx context.Context, projectID uuid.UUID, feedback string) (uuid.UUID, error)
}

func (s *Service) CreateFeedback(ctx context.Context, chatID, messageID int64) (uuid.UUID, error) {
	projectID, err := s.projectByChatIDGetter.GetProjectByChatID(ctx, chatID)
	if err != nil {
		return uuid.Nil, err
	}

	meta := &feedback.CallbackMeta{}
	if err := s.messageMetaScanner.ScanMessageMeta(ctx, chatID, messageID, meta); err != nil {
		return uuid.Nil, err
	}

	parsedMsg, err := message.ParseMessage(meta.Raw, "")
	if err != nil {
		return uuid.Nil, err
	}
	return s.feedbackCreator.NewFeedback(ctx, projectID, parsedMsg.Text)
}

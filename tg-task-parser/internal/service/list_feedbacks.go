package service

import (
	"context"

	"github.com/corray333/tg-task-parser/internal/entities/feedback"
	"github.com/corray333/tg-task-parser/internal/entities/message"
	"github.com/google/uuid"
)

type feedbackLister interface {
	ListFeedbacks(ctx context.Context, projectID uuid.UUID, statuses []feedback.Status) ([]feedback.Feedback, error)
}

type messageMetaSetter interface {
	SetMessageMeta(ctx context.Context, chatID, messageID int64, meta any) error
}

func (s *Service) listActiveFeedbacks(ctx context.Context, projectID uuid.UUID) ([]feedback.Feedback, error) {
	return s.feedbackLister.ListFeedbacks(ctx, projectID, feedback.ActiveStatuses)
}

func (s *Service) RequestActiveFeedbacks(ctx context.Context, chatID int64, messageID int64, msg *message.Message) ([]feedback.Feedback, error) {
	projectID, err := s.projectByChatIDGetter.GetProjectByChatID(ctx, chatID)
	if err != nil {
		return nil, err
	}

	cbm := &feedback.CallbackMeta{
		Raw:       msg.Raw,
		MessageID: messageID,
	}

	if err := s.messageMetaSetter.SetMessageMeta(ctx, chatID, messageID, cbm); err != nil {
		return nil, err
	}

	return s.listActiveFeedbacks(ctx, projectID)
}

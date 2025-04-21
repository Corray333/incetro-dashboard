package service

import (
	"context"

	"github.com/corray333/tg-task-parser/internal/entities/feedback"
	"github.com/corray333/tg-task-parser/internal/entities/message"
	"github.com/google/uuid"
)

type messageMetaScanner interface {
	ScanMessageMeta(ctx context.Context, chatID, messageID int64, meta *feedback.CallbackMeta) error
}

type feedbackAnswerer interface {
	AnswerFeedback(ctx context.Context, feedbackID uuid.UUID, answer string) error
}

func (s *Service) AnswerFeedback(ctx context.Context, chatID, messageID int64, feedbackID uuid.UUID) error {
	meta := &feedback.CallbackMeta{}
	if err := s.messageMetaScanner.ScanMessageMeta(ctx, chatID, messageID, meta); err != nil {
		return err
	}

	parsedMsg, err := message.ParseMessage(meta.Raw, "")
	if err != nil {
		return err
	}

	if err := s.feedbackAnswerer.AnswerFeedback(ctx, feedbackID, parsedMsg.Text); err != nil {
		return err
	}

	return nil
}

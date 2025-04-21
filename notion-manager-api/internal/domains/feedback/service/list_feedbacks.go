package service

import (
	"context"

	"github.com/Corray333/employee_dashboard/internal/domains/feedback/entities/feedback"
)

type feedbacksLister interface {
	ListFeedbacks(ctx context.Context, filter *feedback.Feedback) ([]feedback.Feedback, error)
}

func (s *FeedbackService) ListFeedbacks(ctx context.Context, filter *feedback.Feedback) ([]feedback.Feedback, error) {
	return s.feedbacksLister.ListFeedbacks(ctx, filter)
}

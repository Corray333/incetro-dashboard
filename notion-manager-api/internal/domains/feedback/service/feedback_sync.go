package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/feedback/entities/feedback"
)

type feedbackLastUpdateTimeGetter interface {
	GetFeedbackLastUpdateTime(ctx context.Context) (time.Time, error)
}

type feedbackLastUpdateTimeSetter interface {
	SetFeedbackLastUpdateTime(ctx context.Context, lastUpdateTime time.Time) error
}

type feedbacksRawLister interface {
	ListFeedback(ctx context.Context, lastUpdate time.Time) ([]feedback.Feedback, error)
}

type feedbackSetter interface {
	SetFeedback(ctx context.Context, feedback *feedback.Feedback) error
}

func (s *FeedbackService) updateFeedbacks(ctx context.Context) error {
	slog.Info("Updating feedbacks")
	lastUpdateTime, err := s.feedbackLastUpdateTimeGetter.GetFeedbackLastUpdateTime(ctx)
	if err != nil {
		return err
	}

	feedbacks, err := s.feedbacksRawLister.ListFeedback(ctx, lastUpdateTime)
	if err != nil {
		return err
	}
	if len(feedbacks) == 0 {
		return nil
	}

	lastTime := time.Time{}
	for _, feedback := range feedbacks {
		if err := s.feedbackSetter.SetFeedback(ctx, &feedback); err != nil {
			return err
		}
		fmt.Printf("Feedback %+v\n", feedback)
		if feedback.LastUpdate.After(lastTime) {
			lastTime = feedback.LastUpdate
		}
	}
	if err := s.feedbackLastUpdateTimeSetter.SetFeedbackLastUpdateTime(ctx, lastTime); err != nil {
		return err
	}

	return nil
}

func (s *FeedbackService) FeedbackSync(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		if err := s.updateFeedbacks(ctx); err != nil {
			slog.Error("Notion feedback sync error", "error", err)
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return
		}
	}
}

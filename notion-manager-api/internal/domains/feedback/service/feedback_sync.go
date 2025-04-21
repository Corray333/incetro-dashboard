package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/feedback/entities/feedback"
	"github.com/spf13/viper"
)

type feedbackLastUpdateTimeGetter interface {
	GetFeedbackLastUpdateTime(ctx context.Context) (time.Time, error)
}

type feedbacksRawLister interface {
	ListFeedback(ctx context.Context, lastUpdate time.Time) ([]feedback.Feedback, error)
}

type feedbackSetter interface {
	SetFeedback(ctx context.Context, feedback *feedback.Feedback) error
}

func (s *FeedbackService) updateFeedbacks(ctx context.Context) error {
	lastUpdateTime, err := s.feedbackLastUpdateTimeGetter.GetFeedbackLastUpdateTime(ctx)
	if err != nil {
		return err
	}

	feedbacks, err := s.feedbacksRawLister.ListFeedback(ctx, lastUpdateTime)
	if err != nil {
		return err
	}

	for _, feedback := range feedbacks {
		if err := s.feedbackSetter.SetFeedback(ctx, &feedback); err != nil {
			return err
		}
	}

	return nil
}

func (s *FeedbackService) FeedbackSync(ctx context.Context) {
	if err := s.updateFeedbacks(ctx); err != nil {
		slog.Error("Notion feedback sync error", "error", err)
	}

	ticker := time.NewTicker(viper.GetDuration("notion.sync_interval"))
	for {
		select {
		case <-ticker.C:
			go s.FeedbackSync(ctx)
		case <-ctx.Done():
			return
		}
	}
}

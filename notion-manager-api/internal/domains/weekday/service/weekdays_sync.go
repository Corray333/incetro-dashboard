package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/weekday/entities/weekday"
)

type weekdaysNotionLister interface {
	ListWeekdays(ctx context.Context, lastUpdate time.Time) ([]weekday.Weekday, error)
}

type weekdayLastUpdateTimeGetter interface {
	GetWeekdaysLastUpdateTime(ctx context.Context) (time.Time, error)
}

type weekdaySetter interface {
	SetWeekday(ctx context.Context, weekday *weekday.Weekday) error
}

func (s *WeekdayService) WeekdaysSync(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		if err := s.updateWeekdays(ctx); err != nil {
			slog.Error("Notion weekdays sync error", "error", err)
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return
		}
	}

}

func (s *WeekdayService) updateWeekdays(ctx context.Context) error {
	lastUpdateTime, err := s.weekdayLastUpdateTimeGetter.GetWeekdaysLastUpdateTime(ctx)
	if err != nil {
		return err
	}

	weekdays, err := s.weekdaysNotionLister.ListWeekdays(ctx, lastUpdateTime)
	if err != nil {
		return err
	}
	if len(weekdays) == 0 {
		return nil
	}

	ctx, err = s.transactioner.Begin(ctx)
	if err != nil {
		return err
	}
	defer s.transactioner.Rollback(ctx)

	for _, w := range weekdays {
		if err := s.weekdaySetter.SetWeekday(ctx, &w); err != nil {
			return err
		}
	}
	if err := s.transactioner.Commit(ctx); err != nil {
		return err
	}

	return nil

}

package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/weekday/entities/weekday"
)

type newWeekdayNotifier interface {
	SendWeekendNotification(ctx context.Context, weekday *weekday.Weekday) error
}

type weekdaywLister interface {
	ListWeekdays(ctx context.Context, filter *weekday.Filter) ([]weekday.Weekday, error)
}

type weekdayNotifiedMaker interface {
	MakeWeekdayNotified(ctx context.Context, weekday *weekday.Weekday) error
}

func (s *WeekdayService) StartWeekdaysNotificationWorker(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		if err := s.notifyAboutNewWeekdays(ctx); err != nil {
			slog.Error("failed to notify about new weekdays", "error", err)
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return
		}
	}

}

func (s *WeekdayService) notifyAboutNewWeekdays(ctx context.Context) error {
	notified := false

	weekdays, err := s.weekdaywLister.ListWeekdays(ctx, &weekday.Filter{
		UpdatedAtTo: time.Now().Add(-time.Minute * 5),
		Notified:    &notified,
	})
	if err != nil {
		return err
	}

	fmt.Println(weekdays)
	if len(weekdays) == 0 {
		return nil
	}

	wg := sync.WaitGroup{}

	for _, w := range weekdays {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx := context.Background()
			ctx, err = s.transactioner.Begin(ctx)
			if err != nil {
				return
			}
			defer s.transactioner.Rollback(ctx)

			if err := s.newWeekdayNotifier.SendWeekendNotification(ctx, &w); err != nil {
				return
			}

			if err := s.weekdayNotifiedMaker.MakeWeekdayNotified(ctx, &w); err != nil {
				return
			}

			if err := s.transactioner.Commit(ctx); err != nil {
				return
			}

		}()
	}

	wg.Wait()

	return nil

}

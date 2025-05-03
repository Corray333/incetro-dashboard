package service

import (
	"context"
	"log/slog"
	"time"

	entity_time "github.com/Corray333/employee_dashboard/internal/domains/time/entities/time"
)

type timeOutboxMsgGetter interface {
	GetTimesMsg(ctx context.Context) (times []entity_time.TimeOutboxMsg, err error)
}

type timeWriteOfSentMarker interface {
	MarkTimeWriteOfAsSent(ctx context.Context, time *entity_time.TimeOutboxMsg) error
}

type timeWriteOfNotion interface {
	CreateTimeWriteOf(ctx context.Context, writeOf *entity_time.TimeOutboxMsg) error
}

func (s *TimeService) processTimeWriteOfs(ctx context.Context) error {
	times, err := s.timeOutboxMsgGetter.GetTimesMsg(ctx)
	if err != nil {
		return err
	}

	for _, time := range times {
		if err := s.timeWriteOfNotion.CreateTimeWriteOf(ctx, &time); err != nil {
			return err
		}
		if err := s.timeWriteOfSentMarker.MarkTimeWriteOfAsSent(ctx, &time); err != nil {
			return err
		}
	}

	return nil
}

func (s *TimeService) StartWriteOfOutboxWorker(ctx context.Context) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		if err := s.processTimeWriteOfs(ctx); err != nil {
			slog.Error("Error processing time write of", "error", err)
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return
		}
	}
}

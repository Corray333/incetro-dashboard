package service

import (
	"context"
	"log/slog"
	"time"
	pkg_time "time"

	entity_time "github.com/Corray333/employee_dashboard/internal/domains/time/entities/time"
	"github.com/google/uuid"
)

type timeLastUpdateTimeGetter interface {
	GetTimeLastUpdateTime(ctx context.Context) (time.Time, error)
}

type timeLastUpdateTimeSetter interface {
	SetTimeLastUpdateTime(ctx context.Context, lastUpdateTime time.Time) error
}

type timeRawLister interface {
	ListTimes(ctx context.Context, lastUpdate pkg_time.Time) ([]entity_time.Time, error)
}

type timeSetter interface {
	SetTime(ctx context.Context, time *entity_time.Time) error
}

type timesLister interface {
	ListTimes(ctx context.Context, offset, limit int) ([]entity_time.Time, error)
	ListTimesByProject(ctx context.Context, projectID uuid.UUID, offset int, limit int) ([]entity_time.Time, error)
}

func (s *TimeService) updateTimes(ctx context.Context) error {
	slog.Info("Updating times")
	lastUpdateTime, err := s.timeLastUpdateTimeGetter.GetTimeLastUpdateTime(ctx)
	if err != nil {
		return err
	}

	times, err := s.timeRawLister.ListTimes(ctx, lastUpdateTime)
	if err != nil {
		return err
	}
	if len(times) == 0 {
		return nil
	}

	lastTime := time.Time{}
	for _, time := range times {
		if err := s.timeSetter.SetTime(ctx, &time); err != nil {
			return err
		}
		if time.LastUpdate.After(lastTime) {
			lastTime = time.LastUpdate
		}
	}
	if err := s.timeLastUpdateTimeSetter.SetTimeLastUpdateTime(ctx, lastTime); err != nil {
		return err
	}

	return nil
}

func (s *TimeService) TimeSync(ctx context.Context) {
	ticker := time.NewTicker(20 * time.Second)
	defer ticker.Stop()

	for {
		if err := s.updateTimes(ctx); err != nil {
			slog.Error("Notion time sync error", "error", err)
		}

		select {
		case <-ticker.C:
			continue
		case <-ctx.Done():
			return
		}
	}
}

func (s *TimeService) updateSheets(ctx context.Context) {
	// times, err := postgresRepo.ListTimes(context.Background(), 0, 9000)
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(sheetsRepo.UpdateSheetsTimes(context.Background(), times))

	times, err := s.timesLister.ListTimes(ctx, 0, 20000)
	if err != nil {
		slog.Error("Error getting times", "error", err)
		return
	}

	if err := s.sheetsRepository.UpdateSheetsTimes(ctx, times); err != nil {
		slog.Error("Error updating sheets", "error", err)
		return
	}

	times, err = s.timesLister.ListTimesByProject(ctx, uuid.MustParse("e754753f-491b-4fd5-8913-d1fc51ce2f12"), 0, 20000)
	if err != nil {
		slog.Error("Error getting project temp times", "error", err)
		return
	}

	if err := s.sheetsRepository.UpdateTempSheetsTimes(ctx, times); err != nil {
		slog.Error("Error updating project temp sheets", "error", err)
		return
	}

}

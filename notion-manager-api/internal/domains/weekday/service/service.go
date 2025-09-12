package service

import (
	"context"
	"log/slog"

	"github.com/Corray333/employee_dashboard/internal/domains/weekday/entities/weekday"
	"github.com/Corray333/employee_dashboard/internal/postgres"
	"github.com/spf13/viper"
)

type WeekdayService struct {
	weekdaySetter               weekdaySetter
	weekdayLastUpdateTimeGetter weekdayLastUpdateTimeGetter
	weekdaysNotionLister        weekdaysNotionLister
	transactioner               postgres.Transactioner
	newWeekdayNotifier          newWeekdayNotifier
	weekdaywLister              weekdaywLister
	weekdayNotifiedMaker        weekdayNotifiedMaker
	sheetsRepository            sheetsRepository
}

type postgresRepository interface {
	postgres.Transactioner
	weekdaySetter
	weekdayLastUpdateTimeGetter
	weekdaywLister
	weekdayNotifiedMaker
}
type notionRepository interface {
	weekdaysNotionLister
}

type telegramRepository interface {
	newWeekdayNotifier
}

type option func(*WeekdayService)

type sheetsRepository interface {
	UpdateSheetsWeekdays(ctx context.Context, sheetID string, weekdays []weekday.Weekday) error
}

func (s *WeekdayService) AcceptUpdate(ctx context.Context) {
	go s.updateSheets(ctx)
}

func (s *WeekdayService) updateSheets(ctx context.Context) {

	tasks, err := s.weekdaywLister.ListWeekdays(ctx, &weekday.Filter{})
	if err != nil {
		slog.Error("Error getting tasks", "error", err)
		return
	}

	if err := s.sheetsRepository.UpdateSheetsWeekdays(ctx, viper.GetString("sheets.id"), tasks); err != nil {
		slog.Error("Error updating sheets", "error", err)
		return
	}
}

func NewWeekdayService(opts ...option) *WeekdayService {
	service := &WeekdayService{}

	for _, opt := range opts {
		opt(service)
	}

	return service
}

func WithPostgresRepository(repository postgresRepository) option {
	return func(s *WeekdayService) {
		s.transactioner = repository
		s.weekdaySetter = repository
		s.weekdayLastUpdateTimeGetter = repository
		s.weekdaywLister = repository
		s.weekdayNotifiedMaker = repository
	}
}

func WithSheetsRepository(repository sheetsRepository) option {
	return func(s *WeekdayService) {
		s.sheetsRepository = repository
	}
}

func WithNotionRepository(repository notionRepository) option {
	return func(s *WeekdayService) {
		s.weekdaysNotionLister = repository
	}
}

func WithTelegramRepository(repository telegramRepository) option {
	return func(s *WeekdayService) {
		s.newWeekdayNotifier = repository
	}
}

func (s *WeekdayService) Run() {
	go s.WeekdaysSync(context.Background())
	go s.StartWeekdaysNotificationWorker(context.Background())
}

package service

import (
	"context"

	"github.com/Corray333/employee_dashboard/internal/postgres"
)

type WeekdayService struct {
	weekdaySetter               weekdaySetter
	weekdayLastUpdateTimeGetter weekdayLastUpdateTimeGetter
	weekdaysNotionLister        weekdaysNotionLister
	transactioner               postgres.Transactioner
	newWeekdayNotifier          newWeekdayNotifier
	weekdaywLister              weekdaywLister
	weekdayNotifiedMaker        weekdayNotifiedMaker
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

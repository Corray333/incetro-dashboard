package service

import (
	"context"

	entity_time "github.com/Corray333/employee_dashboard/internal/domains/time/entities/time"
)

type TimeService struct {
	timeLastUpdateTimeGetter
	timeLastUpdateTimeSetter
	timeRawLister
	timeSetter
	sheetsRepository
	timesLister
}

type postgresRepository interface {
	timeLastUpdateTimeGetter
	timeLastUpdateTimeSetter
	timeSetter
	timesLister
}
type notionRepository interface {
	timeRawLister
}

type sheetsRepository interface {
	UpdateSheetsTimes(ctx context.Context, times []entity_time.Time) error
}

type option func(*TimeService)

func NewTaskService(opts ...option) *TimeService {
	service := &TimeService{}

	for _, opt := range opts {
		opt(service)
	}

	return service
}

func (s *TimeService) AcceptUpdate(ctx context.Context) {
	go s.updateSheets(ctx)
}

func WithPostgresRepository(repository postgresRepository) option {
	return func(s *TimeService) {
		s.timeLastUpdateTimeGetter = repository
		s.timeLastUpdateTimeSetter = repository
		s.timeSetter = repository
		s.timesLister = repository
	}
}

func WithNotionRepository(repository notionRepository) option {
	return func(s *TimeService) {
		s.timeRawLister = repository
	}
}

func WithSheetsRepository(repository sheetsRepository) option {
	return func(s *TimeService) {
		s.sheetsRepository = repository
	}
}

func (s *TimeService) Run() {
	go s.TimeSync(context.Background())
}

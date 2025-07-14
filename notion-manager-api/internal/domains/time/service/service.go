package service

import (
	"context"

	entity_time "github.com/Corray333/employee_dashboard/internal/domains/time/entities/time"
	"github.com/google/uuid"
)

type TimeService struct {
	timeLastUpdateTimeGetter timeLastUpdateTimeGetter
	timeLastUpdateTimeSetter timeLastUpdateTimeSetter
	timeRawLister            timeRawLister
	timeSetter               timeSetter
	sheetsRepository         sheetsRepository
	timesLister              timesLister
	timeWriteOfCreater       timeWriteOfCreater
	timeOutboxMsgGetter      timeOutboxMsgGetter
	timeWriteOfSentMarker    timeWriteOfSentMarker
	timeWriteOfNotion        timeWriteOfNotion
	projectsLister           projectsLister
	timeDeleter              timeDeleter
}

type postgresRepository interface {
	timeLastUpdateTimeGetter
	timeLastUpdateTimeSetter
	timeSetter
	timesLister
	timeWriteOfCreater
	timeOutboxMsgGetter
	timeWriteOfSentMarker
	timeDeleter
}
type notionRepository interface {
	timeRawLister
	timeWriteOfNotion
}

type sheetsRepository interface {
	UpdateSheetsTimes(ctx context.Context, sheetID string, times []entity_time.Time) error
}

type option func(*TimeService)

func NewTimeService(opts ...option) *TimeService {
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
		s.timeWriteOfCreater = repository
		s.timeOutboxMsgGetter = repository
		s.timeWriteOfSentMarker = repository
		s.timeDeleter = repository
	}
}

func WithNotionRepository(repository notionRepository) option {
	return func(s *TimeService) {
		s.timeRawLister = repository
		s.timeWriteOfNotion = repository
	}
}

func WithSheetsRepository(repository sheetsRepository) option {
	return func(s *TimeService) {
		s.sheetsRepository = repository
	}
}

func WithProjectRepository(repository projectsLister) option {
	return func(s *TimeService) {
		s.projectsLister = repository
	}
}

func (s *TimeService) Run() {
	go s.TimeSync(context.Background())
	go s.StartWriteOfOutboxWorker(context.Background())
}

type timeDeleter interface {
	DeleteTime(ctx context.Context, timeID uuid.UUID) error
}

// DeleteTime deletes a time entry by its ID
func (s *TimeService) DeleteTime(ctx context.Context, timeID uuid.UUID) error {
	return s.timeDeleter.DeleteTime(ctx, timeID)
}

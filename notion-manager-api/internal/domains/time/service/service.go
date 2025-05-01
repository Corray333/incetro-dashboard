package service

import "context"

type TimeService struct {
	timeLastUpdateTimeGetter
	timeLastUpdateTimeSetter
	timeRawLister
	timeSetter
}

type postgresRepository interface {
	timeLastUpdateTimeGetter
	timeLastUpdateTimeSetter
	timeSetter
}
type notionRepository interface {
	timeRawLister
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
	go s.TimeSync(ctx)
}

func WithPostgresRepository(repository postgresRepository) option {
	return func(s *TimeService) {
		s.timeLastUpdateTimeGetter = repository
		s.timeLastUpdateTimeSetter = repository
		s.timeSetter = repository
	}
}

func WithNotionRepository(repository notionRepository) option {
	return func(s *TimeService) {
		s.timeRawLister = repository
	}
}

func (s *TimeService) Run() {
	go s.TimeSync(context.Background())
}

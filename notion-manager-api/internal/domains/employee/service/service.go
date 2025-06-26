package service

type EmployeeService struct {
	employeeLister employeeLister
}

type postgresRepository interface {
	employeeLister
}
type notionRepository interface {
}

type option func(*EmployeeService)

func NewWeekdayService(opts ...option) *EmployeeService {
	service := &EmployeeService{}

	for _, opt := range opts {
		opt(service)
	}

	return service
}

func WithPostgresRepository(repository postgresRepository) option {
	return func(s *EmployeeService) {
		s.employeeLister = repository
	}
}

func WithNotionRepository(repository notionRepository) option {
	return func(s *EmployeeService) {
	}
}

func (s *EmployeeService) Run() {
	// go s.FeedbackSync(context.Background())
}

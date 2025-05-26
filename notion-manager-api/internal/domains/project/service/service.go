package service

type ProjectService struct {
	projectsLister projectsLister
}

type postgresRepository interface {
	projectsLister
}
type notionRepository interface {
	// feedbacksRawLister
}

type option func(*ProjectService)

func NewProjectService(opts ...option) *ProjectService {
	service := &ProjectService{}

	for _, opt := range opts {
		opt(service)
	}

	return service
}

func WithPostgresRepository(repository postgresRepository) option {
	return func(s *ProjectService) {
		s.projectsLister = repository
	}
}

func WithProjectsLister(lister projectsLister) option {
	return func(s *ProjectService) {
		s.projectsLister = lister
	}
}

func WithNotionRepository(repository notionRepository) option {
	return func(s *ProjectService) {}
}

func (s *ProjectService) Run() {

}

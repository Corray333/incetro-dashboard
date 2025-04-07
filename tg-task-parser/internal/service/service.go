package service

type repository interface {
	taskCreator
}

type Service struct {
	taskCreator taskCreator
}

type option func(*Service)

func New(options ...option) *Service {
	s := &Service{}
	for _, opt := range options {
		opt(s)
	}
	return s
}

func WithTaskCreator(taskCreator taskCreator) option {
	return func(s *Service) {
		s.taskCreator = taskCreator
	}
}

func WithRepository(repository repository) option {
	return func(s *Service) {
		s.taskCreator = repository
	}
}

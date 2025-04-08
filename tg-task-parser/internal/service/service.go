package service

type baseService interface {
	taskCreator
	projectsGetter
}

type repository interface {
	chatToProjectLinker
	projectByChatIDGetter
}

type Service struct {
	taskCreator           taskCreator
	projectsGetter        projectsGetter
	chatToProjectLinker   chatToProjectLinker
	projectByChatIDGetter projectByChatIDGetter
}

type option func(*Service)

func New(options ...option) *Service {
	s := &Service{}
	for _, opt := range options {
		opt(s)
	}
	return s
}

func WithProjectByChatIDGetter(projectByChatIDGetter projectByChatIDGetter) option {
	return func(s *Service) {
		s.projectByChatIDGetter = projectByChatIDGetter
	}
}

func WithProjectsGetter(projectsGetter projectsGetter) option {
	return func(s *Service) {
		s.projectsGetter = projectsGetter
	}
}

func WithTaskCreator(taskCreator taskCreator) option {
	return func(s *Service) {
		s.taskCreator = taskCreator
	}
}

func WithChatToProjectLinker(chatToProjectLinker chatToProjectLinker) option {
	return func(s *Service) {
		s.chatToProjectLinker = chatToProjectLinker
	}
}

func WithBaseService(baseService baseService) option {
	return func(s *Service) {
		s.taskCreator = baseService
		s.projectsGetter = baseService
	}
}

func WithRepository(repository repository) option {
	return func(s *Service) {
		s.chatToProjectLinker = repository
		s.projectByChatIDGetter = repository
	}
}

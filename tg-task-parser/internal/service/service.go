package service

type baseService interface {
	taskCreator
	projectsGetter
	feedbackLister
}

type repository interface {
	chatToProjectLinker
	projectByChatIDGetter
	messageMetaSetter
	messageMetaScanner
}

type notionRepo interface {
	feedbackAnswerer
}

type Service struct {
	taskCreator           taskCreator
	projectsGetter        projectsGetter
	chatToProjectLinker   chatToProjectLinker
	projectByChatIDGetter projectByChatIDGetter
	feedbackLister        feedbackLister
	messageMetaSetter     messageMetaSetter
	messageMetaScanner    messageMetaScanner

	feedbackAnswerer feedbackAnswerer
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

func WithFeedbackLister(feedbackLister feedbackLister) option {
	return func(s *Service) {
		s.feedbackLister = feedbackLister
	}
}

func WithBaseService(baseService baseService) option {
	return func(s *Service) {
		s.taskCreator = baseService
		s.projectsGetter = baseService
		s.feedbackLister = baseService
	}
}

func WithMessageMetaSetter(messageMetaSetter messageMetaSetter) option {
	return func(s *Service) {
		s.messageMetaSetter = messageMetaSetter
	}
}

func WithMessageMetaScanner(messageMetaScanner messageMetaScanner) option {
	return func(s *Service) {
		s.messageMetaScanner = messageMetaScanner
	}
}

func WithRepository(repository repository) option {
	return func(s *Service) {
		s.chatToProjectLinker = repository
		s.projectByChatIDGetter = repository
		s.messageMetaSetter = repository
		s.messageMetaScanner = repository
	}
}

func WithFeedbackAnswerer(feedbackAnswerer feedbackAnswerer) option {
	return func(s *Service) {
		s.feedbackAnswerer = feedbackAnswerer
	}
}

func WithNotionRepo(notionRepo notionRepo) option {
	return func(s *Service) {
		s.feedbackAnswerer = notionRepo
	}
}

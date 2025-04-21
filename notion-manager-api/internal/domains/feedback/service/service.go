package service

import "context"

type FeedbackService struct {
	feedbacksLister
	feedbackSetter
	feedbacksRawLister
	feedbackLastUpdateTimeGetter
}

type postgresRepository interface {
	feedbacksLister
	feedbackSetter
	feedbackLastUpdateTimeGetter
}
type notionRepository interface {
	feedbacksRawLister
}

type option func(*FeedbackService)

func NewTaskService(opts ...option) *FeedbackService {
	service := &FeedbackService{}

	for _, opt := range opts {
		opt(service)
	}

	return service
}

func WithPostgresRepository(repository postgresRepository) option {
	return func(s *FeedbackService) {
		s.feedbacksLister = repository
		s.feedbackLastUpdateTimeGetter = repository
		s.feedbackSetter = repository
	}
}

func WithFeedbacksLister(lister feedbacksLister) option {
	return func(s *FeedbackService) {
		s.feedbacksLister = lister
	}
}

func WithFeedbackSetter(setter feedbackSetter) option {
	return func(s *FeedbackService) {
		s.feedbackSetter = setter
	}
}

func WithFeedbackLastUpdateTimeGetter(getter feedbackLastUpdateTimeGetter) option {
	return func(s *FeedbackService) {
		s.feedbackLastUpdateTimeGetter = getter
	}
}

func WithNotionRepository(repository notionRepository) option {
	return func(s *FeedbackService) {
		s.feedbacksRawLister = repository
	}
}

func WithFeedbacksRawLister(lister feedbacksRawLister) option {
	return func(s *FeedbackService) {
		s.feedbacksRawLister = lister
	}
}

func (s *FeedbackService) Run() {
	go s.FeedbackSync(context.Background())
}

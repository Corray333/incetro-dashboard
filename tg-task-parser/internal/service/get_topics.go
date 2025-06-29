package service

import (
	"context"

	"github.com/corray333/tg-task-parser/internal/entities/topic"
)

type topicsGetter interface {
	GetTopics(ctx context.Context) ([]topic.Topic, error)
}

func (s *Service) GetTopics(ctx context.Context) ([]topic.Topic, error) {
	return s.topicsGetter.GetTopics(ctx)
}

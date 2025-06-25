package service

import (
	"context"

	"github.com/corray333/tg-task-parser/internal/entities/message"
)

type tgMessageSaver interface {
	SaveMessage(ctx context.Context, message message.Message) error
}

func (s *Service) SaveTgMessage(ctx context.Context, msg message.Message) error {
	return s.tgMessageSaver.SaveMessage(ctx, msg)
}

package service

import (
	"context"

	"github.com/google/uuid"
)

type chatToProjectLinker interface {
	LinkChatToProject(ctx context.Context, chatID int64, projectID uuid.UUID) error
}

func (s *Service) LinkChatToProject(ctx context.Context, chatID int64, projectID uuid.UUID) error {
	return s.chatToProjectLinker.LinkChatToProject(ctx, chatID, projectID)
}

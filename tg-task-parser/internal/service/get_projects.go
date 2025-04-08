package service

import (
	"context"

	"github.com/corray333/tg-task-parser/internal/entities/project"
)

type projectsGetter interface {
	GetProjects(ctx context.Context) ([]project.Project, error)
}

func (s *Service) GetProjects(ctx context.Context) ([]project.Project, error) {
	return s.projectsGetter.GetProjects(ctx)
}

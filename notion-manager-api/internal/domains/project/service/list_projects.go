package service

import (
	"context"

	"github.com/Corray333/employee_dashboard/internal/domains/project/entities/project"
)

type projectsLister interface {
	ListProjects(ctx context.Context) ([]project.Project, error)
}

func (s *ProjectService) ListProjects(ctx context.Context) ([]project.Project, error) {
	return s.projectsLister.ListProjects(ctx)
}

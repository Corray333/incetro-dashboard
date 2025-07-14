package service

import (
	"context"

	"github.com/Corray333/employee_dashboard/internal/domains/project/entities/project"
	"github.com/google/uuid"
)

type projectsLister interface {
	ListProjects(ctx context.Context) ([]project.Project, error)
}

func (s *ProjectService) ListProjects(ctx context.Context) ([]project.Project, error) {
	return s.projectsLister.ListProjects(ctx)
}

type projectWithSheetsLister interface {
	ListProjectsWithLinkedSheets(ctx context.Context) ([]project.Project, error)
}

func (s *ProjectService) ListProjectsWithLinkedSheets(ctx context.Context) ([]project.Project, error) {
	return s.projectWithSheetsLister.ListProjectsWithLinkedSheets(ctx)
}

type projectsByIDsGetter interface {
	GetProjectsByIDs(ctx context.Context, projectIDs []uuid.UUID) ([]project.Project, error)
}

func (s *ProjectService) GetProjectsByIDs(ctx context.Context, projectIDs []uuid.UUID) ([]project.Project, error) {
	return s.projectsByIDsGetter.GetProjectsByIDs(ctx, projectIDs)
}

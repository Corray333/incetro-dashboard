package service

import (
	"context"

	"github.com/google/uuid"
)

type ProjectService struct {
	projectsLister          projectsLister
	projectWithSheetsLister projectWithSheetsLister
	projectsByIDsGetter     projectsByIDsGetter

	projectSheetsUpdaters []ProjectSheetsUpdater
}

type ProjectSheetsUpdater interface {
	UpdateProjectSheets(ctx context.Context, projectID uuid.UUID) error
}

type postgresRepository interface {
	projectsLister
	projectWithSheetsLister
	projectsByIDsGetter
}
type notionRepository interface {
	// feedbacksRawLister
}

type option func(*ProjectService)

func NewProjectService(opts ...option) *ProjectService {
	service := &ProjectService{}

	for _, opt := range opts {
		opt(service)
	}

	return service
}

func (s *ProjectService) AddProjectSheetsUpdater(updater ProjectSheetsUpdater) {
	s.projectSheetsUpdaters = append(s.projectSheetsUpdaters, updater)
}

func WithPostgresRepository(repository postgresRepository) option {
	return func(s *ProjectService) {
		s.projectsLister = repository
		s.projectWithSheetsLister = repository
		s.projectsByIDsGetter = repository
	}
}

func WithProjectsLister(lister projectsLister) option {
	return func(s *ProjectService) {
		s.projectsLister = lister
	}
}

func WithNotionRepository(repository notionRepository) option {
	return func(s *ProjectService) {}
}

func (s *ProjectService) Run() {

}

func (s *ProjectService) UpdateProjectSheets(ctx context.Context, projectID uuid.UUID) error {
	for _, updater := range s.projectSheetsUpdaters {
		if err := updater.UpdateProjectSheets(ctx, projectID); err != nil {
			return err
		}
	}
	return nil
}

package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Corray333/employee_dashboard/internal/domains/client/entities/client"
	"github.com/Corray333/employee_dashboard/internal/domains/project/entities/project"
	"github.com/google/uuid"
)

type ProjectService struct {
	projectsLister          projectsLister
	projectWithSheetsLister projectWithSheetsLister
	projectsByIDsGetter     projectsByIDsGetter
	clientsByIDsGetter      clientsByIDsGetter
	sheetsRepository        sheetsRepository

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

type clientsByIDsGetter interface {
	GetClientsByIDs(ctx context.Context, clientIDs []uuid.UUID) ([]client.Client, error)
}

type sheetsRepository interface {
	UpdateSheetsProjects(ctx context.Context, sheetID string, projects []project.Project) error
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

func WithClientService(clientService clientsByIDsGetter) option {
	return func(s *ProjectService) {
		s.clientsByIDsGetter = clientService
	}
}

func WithSheetsRepository(repo sheetsRepository) option {
	return func(s *ProjectService) {
		s.sheetsRepository = repo
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

func (s *ProjectService) AcceptUpdate(ctx context.Context) {
	go s.UpdateSheets(ctx)
}

func (s *ProjectService) UpdateProjectSheets(ctx context.Context, projectID uuid.UUID) error {
	for _, updater := range s.projectSheetsUpdaters {
		if err := updater.UpdateProjectSheets(ctx, projectID); err != nil {
			return err
		}
	}
	return nil
}

func (s *ProjectService) UpdateSheets(ctx context.Context) error {
	projects, err := s.projectsLister.ListProjects(ctx)
	if err != nil {
		return err
	}

	// Populate client data for each project
	for i := range projects {
		if projects[i].ClientID != nil {
			fmt.Println("Project client: ", projects[i].ClientID)
			clients, err := s.clientsByIDsGetter.GetClientsByIDs(ctx, []uuid.UUID{*projects[i].ClientID})
			if err != nil {
				slog.Error("Error getting client for project", "projectID", projects[i].ID, "clientID", *projects[i].ClientID, "error", err)
				continue
			}
			if len(clients) > 0 {
				projects[i].Client = &clients[0]
			}
		}
	}

	return s.sheetsRepository.UpdateSheetsProjects(ctx, "", projects)
}

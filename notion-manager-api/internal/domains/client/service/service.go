package service

import (
	"context"

	"github.com/Corray333/employee_dashboard/internal/domains/project/entities/project"
	"github.com/Corray333/employee_dashboard/internal/postgres"
	"github.com/google/uuid"
)

type ClientService struct {
	clientSetter               clientSetter
	clientLastUpdateTimeGetter clientLastUpdateTimeGetter
	clientsNotionLister        clientsNotionLister
	transactioner              postgres.Transactioner
	clientLister               clientLister
	sheetsClientsUpdater       sheetsClientsUpdater
	projectsByIDsGetter        projectsByIDsGetter
}

type postgresRepository interface {
	postgres.Transactioner
	clientSetter
	clientLastUpdateTimeGetter
	clientLister
}

type notionRepository interface {
	clientsNotionLister
}

type sheetsRepository interface {
	sheetsClientsUpdater
}

type projectsByIDsGetter interface {
	GetProjectsByIDs(ctx context.Context, projectIDs []uuid.UUID) ([]project.Project, error)
}

type option func(*ClientService)

func NewClientService(opts ...option) *ClientService {
	service := &ClientService{}

	for _, opt := range opts {
		opt(service)
	}

	return service
}

func WithPostgresRepository(repository postgresRepository) option {
	return func(s *ClientService) {
		s.transactioner = repository
		s.clientSetter = repository
		s.clientLastUpdateTimeGetter = repository
		s.clientLister = repository
	}
}

func WithNotionRepository(repository notionRepository) option {
	return func(s *ClientService) {
		s.clientsNotionLister = repository
	}
}

func WithSheetsRepository(repository sheetsRepository) option {
	return func(s *ClientService) {
		s.sheetsClientsUpdater = repository
	}
}

func WithProjectService(projectService projectsByIDsGetter) option {
	return func(s *ClientService) {
		s.projectsByIDsGetter = projectService
	}
}

func (s *ClientService) Run() {
	go s.ClientsSync(context.Background())
}

func (s *ClientService) AcceptUpdate(ctx context.Context) {
	go s.UpdateSheets(ctx)
}
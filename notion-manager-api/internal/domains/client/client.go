package client

import (
	notion_repo "github.com/Corray333/employee_dashboard/internal/domains/client/repositories/notion"
	postgres_repo "github.com/Corray333/employee_dashboard/internal/domains/client/repositories/postgres"
	sheets_repo "github.com/Corray333/employee_dashboard/internal/domains/client/repositories/sheets"
	"github.com/Corray333/employee_dashboard/internal/domains/client/service"
	project_service "github.com/Corray333/employee_dashboard/internal/domains/project/service"
	"github.com/Corray333/employee_dashboard/internal/postgres"
	gsheets "github.com/Corray333/employee_dashboard/internal/sheets"
	notion "github.com/Corray333/employee_dashboard/pkg/notion/v2"
)

type ClientController struct {
	postgresRepo *postgres_repo.ClientPostgresRepository
	notionRepo   *notion_repo.ClientNotionRepository
	sheetsRepo   *sheets_repo.ClientSheetsRepository
	service      *service.ClientService
	// transport    *transport.ClientTransport
}

func NewClientController(store *postgres.PostgresClient, notionClient *notion.Client, sheetsClient *gsheets.Client, projectService *project_service.ProjectService) *ClientController {
	postgresRepo := postgres_repo.NewClientPostgresRepository(store)
	notionRepo := notion_repo.NewClientNotionRepository(notionClient)
	sheetsRepo := sheets_repo.NewClientSheetsRepository(sheetsClient)

	if projectService != nil {
		service := service.NewClientService(
			service.WithPostgresRepository(postgresRepo),
			service.WithNotionRepository(notionRepo),
			service.WithSheetsRepository(sheetsRepo),
			service.WithProjectService(projectService),
		)

		return &ClientController{
			postgresRepo: postgresRepo,
			notionRepo:   notionRepo,
			sheetsRepo:   sheetsRepo,
			service:      service,
		}
	} else {
		service := service.NewClientService(
			service.WithPostgresRepository(postgresRepo),
			service.WithNotionRepository(notionRepo),
			service.WithSheetsRepository(sheetsRepo),
		)

		return &ClientController{
			postgresRepo: postgresRepo,
			notionRepo:   notionRepo,
			sheetsRepo:   sheetsRepo,
			service:      service,
		}
	}
}

func (c *ClientController) Build() {
	// c.transport.RegisterRoutes()
}

func (c *ClientController) Run() {
	c.service.Run()
}

func (c *ClientController) GetService() *service.ClientService {
	return c.service
}
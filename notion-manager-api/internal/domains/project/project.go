package project

import (
	notion_repo "github.com/Corray333/employee_dashboard/internal/domains/project/repositories/notion"
	postgres_repo "github.com/Corray333/employee_dashboard/internal/domains/project/repositories/postgres"
	sheets_repo "github.com/Corray333/employee_dashboard/internal/domains/project/repositories/sheets"
	"github.com/Corray333/employee_dashboard/internal/domains/project/service"
	"github.com/Corray333/employee_dashboard/internal/domains/project/transport"
	client_service "github.com/Corray333/employee_dashboard/internal/domains/client/service"
	"github.com/Corray333/employee_dashboard/internal/postgres"
	gsheets "github.com/Corray333/employee_dashboard/internal/sheets"
	notion "github.com/Corray333/employee_dashboard/pkg/notion/v2"
	"github.com/go-chi/chi/v5"
)

type ProjectController struct {
	postgresRepo *postgres_repo.ProjectPostgresRepository
	notionRepo   *notion_repo.ProjectNotionRepository
	service      *service.ProjectService
	transport    *transport.ProjectTransport
}

func NewProjectController(router *chi.Mux, store *postgres.PostgresClient, notionClient *notion.Client, sheetsClient *gsheets.Client, clientService *client_service.ClientService) *ProjectController {

	postgresRepo := postgres_repo.NewProjectPostgresRepository(store)
	notionRepo := notion_repo.NewProjectNotionRepository(notionClient)
	sheetsRepo := sheets_repo.NewProjectSheetsRepository(sheetsClient)

	service := service.NewProjectService(
		service.WithPostgresRepository(postgresRepo),
		service.WithNotionRepository(notionRepo),
		service.WithSheetsRepository(sheetsRepo),
		service.WithClientService(clientService),
	)

	transport := transport.NewProjectTransport(router, service)

	return &ProjectController{
		postgresRepo: postgresRepo,
		notionRepo:   notionRepo,

		service:   service,
		transport: transport,
	}
}

func (c *ProjectController) AddProjectSheetsUpdater(updater service.ProjectSheetsUpdater) {
	c.service.AddProjectSheetsUpdater(updater)
}

func (c *ProjectController) Build() {
	c.transport.RegisterRoutes()
}

func (c *ProjectController) Run() {
	// c.service.Run()
}

func (c *ProjectController) GetService() *service.ProjectService {
	return c.service
}

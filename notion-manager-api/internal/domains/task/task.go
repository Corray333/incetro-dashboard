package task

import (
	notion_repo "github.com/Corray333/employee_dashboard/internal/domains/task/repositories/notion"
	postgres_repo "github.com/Corray333/employee_dashboard/internal/domains/task/repositories/postgres"
	sheets_repo "github.com/Corray333/employee_dashboard/internal/domains/task/repositories/sheets"
	"github.com/Corray333/employee_dashboard/internal/domains/task/service"
	"github.com/Corray333/employee_dashboard/internal/domains/task/transport"
	"github.com/Corray333/employee_dashboard/internal/postgres"
	gsheets "github.com/Corray333/employee_dashboard/internal/sheets"
	notion "github.com/Corray333/employee_dashboard/pkg/notion/v2"
	"github.com/go-chi/chi/v5"
)

type TaskController struct {
	postgresRepo *postgres_repo.TaskPostgresRepository
	notionRepo   *notion_repo.TaskNotionRepository
	service      *service.TaskService
	transport    *transport.TaskTransport
}

func NewTaskController(router *chi.Mux, store *postgres.PostgresClient, notionClient *notion.Client, sheetsClient *gsheets.Client) *TaskController {

	postgresRepo := postgres_repo.NewTaskPostgresRepository(store)
	notionRepo := notion_repo.NewTaskNotionRepository(notionClient)
	sheetsRepo := sheets_repo.NewTaskSheetsRepository(sheetsClient)

	service := service.NewTaskService(service.WithPostgresRepository(postgresRepo), service.WithNotionRepository(notionRepo), service.WithSheetsRepository(sheetsRepo))

	transport := transport.NewTaskTransport(router, service)

	return &TaskController{
		postgresRepo: postgresRepo,
		notionRepo:   notionRepo,

		service:   service,
		transport: transport,
	}
}

func (c *TaskController) Build() {
	c.transport.RegisterRoutes()
}

func (c *TaskController) Run() {
	c.service.Run()
}

func (c *TaskController) GetService() *service.TaskService {
	return c.service
}

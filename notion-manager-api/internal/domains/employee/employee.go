package employee

import (
	notion_repo "github.com/Corray333/employee_dashboard/internal/domains/employee/repositories/notion"
	postgres_repo "github.com/Corray333/employee_dashboard/internal/domains/employee/repositories/postgres"
	"github.com/Corray333/employee_dashboard/internal/domains/employee/service"
	"github.com/Corray333/employee_dashboard/internal/postgres"
	notion "github.com/Corray333/employee_dashboard/pkg/notion/v2"
)

type EmployeeController struct {
	postgresRepo *postgres_repo.EmployeePostgresRepository
	notionRepo   *notion_repo.EmployeeNotionRepository
	service      *service.EmployeeService
	// transport    *transport.EmployeeTransport
}

func NewEmployeeController(store *postgres.PostgresClient, notionClient *notion.Client) *EmployeeController {

	postgresRepo := postgres_repo.NewEmployeePostgresRepository(store)
	notionRepo := notion_repo.NewWeekdayNotionRepository(notionClient)

	service := service.NewWeekdayService(service.WithPostgresRepository(postgresRepo), service.WithNotionRepository(notionRepo))

	// transport := transport.NewWeekdayTransport(grpcServer, service)

	return &EmployeeController{
		postgresRepo: postgresRepo,
		notionRepo:   notionRepo,

		service: service,
	}
}

func (c *EmployeeController) GetService() *service.EmployeeService {
	return c.service
}

func (c *EmployeeController) Build() {
	// c.transport.RegisterRoutes()
}

func (c *EmployeeController) Run() {
	c.service.Run()
}

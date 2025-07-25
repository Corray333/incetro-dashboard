package weekday

import (
	employee_service "github.com/Corray333/employee_dashboard/internal/domains/employee/service"
	notion_repo "github.com/Corray333/employee_dashboard/internal/domains/weekday/repositories/notion"
	postgres_repo "github.com/Corray333/employee_dashboard/internal/domains/weekday/repositories/postgres"
	tg_repo "github.com/Corray333/employee_dashboard/internal/domains/weekday/repositories/tg"
	"github.com/Corray333/employee_dashboard/internal/domains/weekday/service"
	"github.com/Corray333/employee_dashboard/internal/postgres"
	"github.com/Corray333/employee_dashboard/internal/telegram"
	notion "github.com/Corray333/employee_dashboard/pkg/notion/v2"
)

type WeekdayController struct {
	postgresRepo *postgres_repo.WeekdayPostgresRepository
	notionRepo   *notion_repo.WeekdayNotionRepository
	service      *service.WeekdayService
	// transport    *transport.WeekdayTransport
}

func NewWeekdayController(store *postgres.PostgresClient, notionClient *notion.Client, tgClient *telegram.TelegramClient, userGetter *employee_service.EmployeeService) *WeekdayController {

	postgresRepo := postgres_repo.NewWeekdayPostgresRepository(store, userGetter)
	notionRepo := notion_repo.NewWeekdayNotionRepository(notionClient)
	tgRepo := tg_repo.NewWeekdayTelegramRepository(tgClient)

	service := service.NewWeekdayService(service.WithPostgresRepository(postgresRepo), service.WithNotionRepository(notionRepo), service.WithTelegramRepository(tgRepo))

	// transport := transport.NewWeekdayTransport(grpcServer, service)

	return &WeekdayController{
		postgresRepo: postgresRepo,
		notionRepo:   notionRepo,

		service: service,
	}
}

func (c *WeekdayController) Build() {
	// c.transport.RegisterRoutes()
}

func (c *WeekdayController) Run() {
	c.service.Run()
}

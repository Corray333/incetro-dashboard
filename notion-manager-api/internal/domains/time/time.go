package feedback

import (
	"context"
	"fmt"

	notion_repo "github.com/Corray333/employee_dashboard/internal/domains/time/repositories/notion"
	postgres_repo "github.com/Corray333/employee_dashboard/internal/domains/time/repositories/postgres"
	"github.com/Corray333/employee_dashboard/internal/domains/time/repositories/sheets"
	"github.com/Corray333/employee_dashboard/internal/domains/time/service"
	"github.com/Corray333/employee_dashboard/internal/domains/time/transport"
	"github.com/Corray333/employee_dashboard/internal/postgres"
	gsheets "github.com/Corray333/employee_dashboard/internal/sheets"
	notion "github.com/Corray333/employee_dashboard/pkg/notion/v2"
	"google.golang.org/grpc"
)

type TimeController struct {
	postgresRepo *postgres_repo.TimePostgresRepository
	notionRepo   *notion_repo.TimeNotionRepository
	service      *service.TimeService
	transport    *transport.TimeTransport
}

func NewTimeController(grpcServer *grpc.Server, store *postgres.PostgresClient, notionClient *notion.Client, sheetsClient *gsheets.Client) *TimeController {

	postgresRepo := postgres_repo.NewTimePostgresRepository(store)
	notionRepo := notion_repo.NewTimeNotionRepository(notionClient)
	sheetsRepo := sheets.NewTimeSheetsRepository(sheetsClient)
	times, err := postgresRepo.ListTimes(context.Background(), 0, 9000)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(sheetsRepo.UpdateSheetsTimes(context.Background(), times))

	service := service.NewTaskService(service.WithPostgresRepository(postgresRepo), service.WithNotionRepository(notionRepo))

	transport := transport.NewTimeTransport(grpcServer, service)

	return &TimeController{
		postgresRepo: postgresRepo,
		notionRepo:   notionRepo,

		service:   service,
		transport: transport,
	}
}

func (c *TimeController) Build() {
	// c.transport.RegisterRoutes()
}

func (c *TimeController) Run() {
	c.service.Run()
}

func (c *TimeController) GetService() *service.TimeService {
	return c.service
}

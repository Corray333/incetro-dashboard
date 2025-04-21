package feedback

import (
	notion_repo "github.com/Corray333/employee_dashboard/internal/domains/feedback/repositories/notion"
	postgres_repo "github.com/Corray333/employee_dashboard/internal/domains/feedback/repositories/postgres"
	"github.com/Corray333/employee_dashboard/internal/domains/feedback/service"
	"github.com/Corray333/employee_dashboard/internal/domains/feedback/transport"
	"github.com/Corray333/employee_dashboard/internal/postgres"
	notion "github.com/Corray333/employee_dashboard/pkg/notion/v2"
	"google.golang.org/grpc"
)

type FeedbackController struct {
	postgresRepo *postgres_repo.FeedbackPostgresRepository
	notionRepo   *notion_repo.FeedbackNotionRepository
	service      *service.FeedbackService
	transport    *transport.FeedbackTransport
}

func NewFeedbackController(grpcServer *grpc.Server, store *postgres.PostgresClient, notionClient *notion.Client) *FeedbackController {

	postgresRepo := postgres_repo.NewFeedbackPostgresRepository(store)
	notionRepo := notion_repo.NewFeedbackNotionRepository(notionClient)

	service := service.NewTaskService(service.WithPostgresRepository(postgresRepo), service.WithNotionRepository(notionRepo))

	transport := transport.NewFeedbackTransport(grpcServer, service)

	return &FeedbackController{
		postgresRepo: postgresRepo,
		notionRepo:   notionRepo,

		service:   service,
		transport: transport,
	}
}

func (c *FeedbackController) Build() {
	// c.transport.RegisterRoutes()
}

func (c *FeedbackController) Run() {
	c.service.Run()
}

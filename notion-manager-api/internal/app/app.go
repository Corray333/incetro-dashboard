package app

import (
	"github.com/Corray333/employee_dashboard/internal/domains/feedback"
	"github.com/Corray333/employee_dashboard/internal/domains/project"
	"github.com/Corray333/employee_dashboard/internal/domains/task"
	time "github.com/Corray333/employee_dashboard/internal/domains/time"
	"github.com/Corray333/employee_dashboard/internal/external"
	"github.com/Corray333/employee_dashboard/internal/postgres"
	"github.com/Corray333/employee_dashboard/internal/repositories"
	"github.com/Corray333/employee_dashboard/internal/service"
	gsheets "github.com/Corray333/employee_dashboard/internal/sheets"
	"github.com/Corray333/employee_dashboard/internal/transport"
	notion "github.com/Corray333/employee_dashboard/pkg/notion/v2"
	"google.golang.org/grpc"
)

type app struct {
	store     *repositories.Storage
	service   *service.Service
	transport *transport.Transport

	grpcServer *grpc.Server

	controllers []controller
}

type controller interface {
	Build()
	Run()
}

func New() *app {

	app := &app{}

	router := transport.NewRouter()
	grpcServer := grpc.NewServer()
	app.grpcServer = grpcServer

	store := postgres.New()
	notionClient := notion.NewClient()
	sheetsClient := gsheets.NewSheetsClient()

	feedbackController := feedback.NewFeedbackController(grpcServer, store, notionClient)
	app.controllers = append(app.controllers, feedbackController)

	projectController := project.NewProjectController(router, store, notionClient, sheetsClient)
	app.controllers = append(app.controllers, projectController)

	timeController := time.NewTimeController(router, store, notionClient, sheetsClient, projectController.GetService())
	app.controllers = append(app.controllers, timeController)

	taskController := task.NewTaskController(router, store, notionClient, sheetsClient, projectController.GetService())
	app.controllers = append(app.controllers, taskController)
	projectController.AddProjectSheetsUpdater(taskController.GetService())
	projectController.AddProjectSheetsUpdater(timeController.GetService())

	storage := repositories.New()
	external := external.New()
	service := service.New(storage, external)
	service.AddUpdateSubscriber(timeController.GetService())
	service.AddUpdateSubscriber(taskController.GetService())

	transport := transport.New(router, service)
	transport.RegisterRoutes()

	app.store = storage
	app.service = service
	app.transport = transport

	for _, c := range app.controllers {
		c.Build()
	}

	return app
}

func (app *app) Run() {
	go app.service.Run()
	for _, c := range app.controllers {
		go c.Run()
	}
	// go func() {
	// 	listener, err := net.Listen("tcp", ":50051")
	// 	if err != nil {
	// 		slog.Error("Failed to listen", "error", err)
	// 		panic(err)
	// 	}
	// 	slog.Info("Starting gRPC server")
	// 	if err := app.grpcServer.Serve(listener); err != nil {
	// 		slog.Error("Failed to serve", "error", err)
	// 		panic(err)
	// 	}
	// }()
	app.transport.Run()
}

func (app *app) Init() *app {
	for _, c := range app.controllers {
		c.Build()
	}
	return app
}

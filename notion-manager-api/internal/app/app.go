package app

import (
	"os"

	"github.com/Corray333/employee_dashboard/internal/domains/client"
	"github.com/Corray333/employee_dashboard/internal/domains/employee"
	"github.com/Corray333/employee_dashboard/internal/domains/feedback"
	"github.com/Corray333/employee_dashboard/internal/domains/project"
	"github.com/Corray333/employee_dashboard/internal/domains/task"
	time "github.com/Corray333/employee_dashboard/internal/domains/time"
	"github.com/Corray333/employee_dashboard/internal/domains/weekday"
	"github.com/Corray333/employee_dashboard/internal/external"
	"github.com/Corray333/employee_dashboard/internal/postgres"
	"github.com/Corray333/employee_dashboard/internal/repositories"
	"github.com/Corray333/employee_dashboard/internal/service"
	gsheets "github.com/Corray333/employee_dashboard/internal/sheets"
	"github.com/Corray333/employee_dashboard/internal/telegram"
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
	telegramClient := telegram.NewTelegramClient(os.Getenv("BOT_TOKEN"))

	feedbackController := feedback.NewFeedbackController(grpcServer, store, notionClient)
	app.controllers = append(app.controllers, feedbackController)

	employeeController := employee.NewEmployeeController(store, notionClient)
	app.controllers = append(app.controllers, employeeController)

	// Create client controller first (without project service dependency)
	clientController := client.NewClientController(store, notionClient, sheetsClient, nil)
	app.controllers = append(app.controllers, clientController)

	projectController := project.NewProjectController(router, store, notionClient, sheetsClient, clientController.GetService())
	app.controllers = append(app.controllers, projectController)

	timeController := time.NewTimeController(router, store, notionClient, sheetsClient, projectController.GetService())
	app.controllers = append(app.controllers, timeController)

	taskController := task.NewTaskController(router, store, notionClient, sheetsClient, projectController.GetService())
	app.controllers = append(app.controllers, taskController)

	weekdayController := weekday.NewWeekdayController(store, notionClient, telegramClient, employeeController.GetService())
	app.controllers = append(app.controllers, weekdayController)

	// Update client controller with project service after project controller is created
	clientController = client.NewClientController(store, notionClient, sheetsClient, projectController.GetService())
	app.controllers[len(app.controllers)-4] = clientController

	projectController.AddProjectSheetsUpdater(taskController.GetService())
	projectController.AddProjectSheetsUpdater(timeController.GetService())

	storage := repositories.New()
	external := external.New()
	service := service.New(storage, external)
	service.AddUpdateSubscriber(timeController.GetService())
	service.AddUpdateSubscriber(taskController.GetService())

	// Set task and time services for deletion operations
	service.SetTaskService(taskController.GetService())
	service.SetTimeService(timeController.GetService())
	service.AddUpdateSubscriber(clientController.GetService())

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

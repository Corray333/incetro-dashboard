package app

import (
	"github.com/Corray333/employee_dashboard/internal/external"
	"github.com/Corray333/employee_dashboard/internal/repositories"
	"github.com/Corray333/employee_dashboard/internal/service"
	"github.com/Corray333/employee_dashboard/internal/transport"
)

type app struct {
	store     *repositories.Storage
	service   *service.Service
	transport *transport.Transport
}

func New() *app {

	storage := repositories.New()
	external := external.New()
	// people, err := storage.GetEmployees()
	// if err != nil {
	// 	panic(err)
	// }
	// projects, err := storage.GetProjects("")
	// if err != nil {
	// 	panic(err)
	// }
	// srv, err := external.NewSheetsClient()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(external.UpdatePeopleSheet(srv, people))
	// fmt.Println(external.UpdateProjectsSheet(srv, projects))
	service := service.New(storage, external)

	transport := transport.New(service)
	transport.RegisterRoutes()

	app := &app{
		store:     storage,
		service:   service,
		transport: transport,
	}

	return app
}

func (app *app) Run() {
	go app.service.Run()
	app.transport.Run()
}

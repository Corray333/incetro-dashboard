package app

import (
	"github.com/corray333/tg-task-parser/internal/repositories"
	"github.com/corray333/tg-task-parser/internal/service"
	"github.com/corray333/tg-task-parser/internal/transport"
)

type app struct {
	store     *repositories.Storage
	service   *service.Service
	transport *transport.Transport
}

func New() *app {

	storage := repositories.New()
	service := service.New(storage)

	transport := transport.New(service)

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

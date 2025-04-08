package app

import (
	"github.com/corray333/tg-task-parser/internal/repositories/base_service"
	"github.com/corray333/tg-task-parser/internal/repositories/storage"
	"github.com/corray333/tg-task-parser/internal/service"
	"github.com/corray333/tg-task-parser/internal/transport"
)

type app struct {
	baseService *base_service.BaseService
	service     *service.Service
	transport   *transport.Transport
}

func New() *app {

	baseService := base_service.NewBaseService()
	repository := storage.NewRepository()
	service := service.New(service.WithBaseService(baseService), service.WithRepository(repository))

	transport := transport.New(service)

	app := &app{
		baseService: baseService,
		service:     service,
		transport:   transport,
	}

	return app
}

func (app *app) Run() {
	app.transport.Run()
}

package app

import (
	"github.com/corray333/tg-task-parser/internal/repositories/base_service"
	"github.com/corray333/tg-task-parser/internal/repositories/notion"
	"github.com/corray333/tg-task-parser/internal/repositories/storage"
	"github.com/corray333/tg-task-parser/internal/service"
	"github.com/corray333/tg-task-parser/internal/transport/incetro_bot"
	"github.com/corray333/tg-task-parser/internal/transport/project_bot"
	notion_api "github.com/corray333/tg-task-parser/pkg/notion/v2"
)

type app struct {
	baseService *base_service.BaseService
	service     *service.Service
	incetro_bot *incetro_bot.IncetroTelegramBot
	project_bot *project_bot.ProjectBot
}

func New() *app {

	baseService := base_service.NewBaseService()
	repository := storage.NewRepository()
	notionClient := notion_api.NewClient()
	notionRepo := notion.NewNotionRepository(notionClient)
	service := service.New(service.WithBaseService(baseService), service.WithRepository(repository), service.WithNotionRepo(notionRepo))

	incetroBotTransport := incetro_bot.NewIncetroBot(service)
	projectTransport := project_bot.NewProjectBot(service)

	app := &app{
		baseService: baseService,
		service:     service,
		incetro_bot: incetroBotTransport,
		project_bot: projectTransport,
	}

	return app
}

func (app *app) Run() {
	app.incetro_bot.Run()
}

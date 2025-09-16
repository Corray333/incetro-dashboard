package app

import (
	"log/slog"
	"os"

	"github.com/corray333/tg-task-parser/internal/repositories/base_service"
	"github.com/corray333/tg-task-parser/internal/repositories/notion"
	"github.com/corray333/tg-task-parser/internal/repositories/openaiclient"
	"github.com/corray333/tg-task-parser/internal/repositories/storage"
	"github.com/corray333/tg-task-parser/internal/repositories/temp_storage"
	"github.com/corray333/tg-task-parser/internal/repositories/tg_repository"
	"github.com/corray333/tg-task-parser/internal/repositories/yatracker"
	"github.com/corray333/tg-task-parser/internal/service"
	"github.com/corray333/tg-task-parser/internal/transport/cron"
	"github.com/corray333/tg-task-parser/internal/transport/incetro_bot"
	"github.com/corray333/tg-task-parser/internal/transport/project_bot"
	notion_api "github.com/corray333/tg-task-parser/pkg/notion/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type app struct {
	baseService *base_service.BaseService
	service     *service.Service
	incetro_bot *incetro_bot.IncetroTelegramBot
	project_bot *project_bot.ProjectBot
	cronService *cron.CronService
}

func New() *app {

	baseService := base_service.NewBaseService()
	repository := storage.NewRepository()
	yaTrackerRepo := yatracker.NewYaTrackerRepository()
	// fmt.Println(yaTrackerRepo.CreateTas(context.Background(), &task.Task{
	// 	Text: "Test task",
	// }))
	// fmt.Println(yaTrackerRepo.SearchTasksByName(context.Background(), &task.Task{
	// 	Title: "Test task",
	// }))
	openaiRepo, err := openaiclient.NewOpenAIRepository()
	if err != nil {
		slog.Error("Failed to create openai repository", "error", err)
		panic(err)
	}

	// openaiRepo.ParseMessage(context.Background(), `#задача
	// Прошу еще посмотреть возможность подключения мобильной метрики от Яндекса для нашего приложения.

	// Так как судя по инструменту, он может полностью покрыть запросы маркетинга по сбору метрик.

	// Сайт app метрики - https://appmetrica.yandex.ru/about
	// Документация - https://appmetrica.yandex.ru/docs/ru/
	// 1. Ознакомиться с сайтом
	// 2. Проверить документацию`)

	notionClient := notion_api.NewClient()
	notionRepo := notion.NewNotionRepository(notionClient)

	// Инициализируем Telegram Bot API для отправки сообщений
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		slog.Error("Failed to create telegram bot", "error", err)
		panic(err)
	}
	tgRepo := tg_repository.NewTgRepository(bot)

	temp_storage := temp_storage.NewTempStorage()

	service := service.New(
		service.WithBaseService(baseService),
		service.WithRepository(repository),
		service.WithNotionRepo(notionRepo),
		service.WithTgRepo(tgRepo),
		service.WithYaTrackerRepo(yaTrackerRepo),
		service.WithTaskMsgParser(openaiRepo),
		service.WithTempStorageRepo(temp_storage),
	)
	// fmt.Println(service.SendIncorrectTimeNotifications(context.Background()))

	incetroBotTransport := incetro_bot.NewIncetroBot(service)
	projectTransport := project_bot.NewProjectBot(service)

	// Инициализируем cron сервис
	cronService := cron.NewCronService(service)

	app := &app{
		baseService: baseService,
		service:     service,
		incetro_bot: incetroBotTransport,
		project_bot: projectTransport,
		cronService: cronService,
	}

	return app
}

func (app *app) Run() {
	go app.incetro_bot.Run()
	go app.project_bot.Run()
	// Запускаем cron сервис
	if err := app.cronService.Start(); err != nil {
		slog.Error("Failed to start cron service", "error", err)
	}
	select {}
}

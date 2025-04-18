package transport

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/PaulSonOfLars/gotgbot/v2/ext/handlers"

	"github.com/corray333/tg-task-parser/internal/entities/project"
	"github.com/google/uuid"
)

type Transport struct {
	service    service
	bot        *gotgbot.Bot
	dispatcher *ext.Dispatcher
	updater    *ext.Updater
}

type service interface {
	CreateTask(ctx context.Context, chatID int64, message string, replyMessage string) (string, error)
	GetProjects(ctx context.Context) ([]project.Project, error)
	LinkChatToProject(ctx context.Context, chatID int64, projectID uuid.UUID) error
}

func New(service service) *Transport {
	token := os.Getenv("TASK_PARSER_BOT_TOKEN")
	bot, err := gotgbot.NewBot(token, nil)
	if err != nil {
		panic("failed to create bot: " + err.Error())
	}

	dispatcher := ext.NewDispatcher(&ext.DispatcherOpts{
		Error: func(bot *gotgbot.Bot, ctx *ext.Context, err error) ext.DispatcherAction {
			slog.Error("Handler error", "error", err)
			return ext.DispatcherActionNoop
		},
	})

	updater := ext.NewUpdater(dispatcher, nil)

	tr := &Transport{
		service:    service,
		bot:        bot,
		dispatcher: dispatcher,
		updater:    updater,
	}

	tr.registerHandlers()

	return tr
}

func (t *Transport) registerHandlers() {
	// Хендлер сообщений
	t.dispatcher.AddHandler(handlers.NewMessage(nil, func(bot *gotgbot.Bot, ctx *ext.Context) error {
		msg := ctx.EffectiveMessage
		if msg == nil {
			return nil
		}

		if msg.NewChatMembers != nil {
			for _, member := range msg.NewChatMembers {
				if member.Id == bot.User.Id {
					return t.handleBotAddedToChat(bot, ctx)
				}
			}
			return nil
		}

		var reply string
		if msg.ReplyToMessage != nil {
			reply = msg.ReplyToMessage.Text
		}

		pageID, err := t.service.CreateTask(context.Background(), msg.Chat.Id, msg.Text, reply)
		if err != nil {
			slog.Error("Error creating task", "error", err)
			_, _ = msg.Reply(bot, "Не удалось создать задачу", nil)
			return nil
		}

		if pageID == "" {
			return nil
		}

		_, err = msg.Reply(bot, fmt.Sprintf("Задача создана: https://notion.so/%s", strings.ReplaceAll(pageID, "-", "")), nil)
		if err != nil {
			slog.Error("Error sending confirmation", "error", err)
		}
		return nil
	}))

	// Хендлер callback-кнопок
	t.dispatcher.AddHandler(handlers.NewCallback(nil, func(bot *gotgbot.Bot, ctx *ext.Context) error {
		cb := ctx.CallbackQuery
		projectID, err := uuid.Parse(cb.Data)
		if err != nil {
			slog.Error("Invalid project UUID", "data", cb.Data)
			return nil
		}

		err = t.service.LinkChatToProject(context.Background(), cb.Message.GetChat().Id, projectID)
		if err != nil {
			slog.Error("Error linking chat to project", "error", err)
			return nil
		}

		_, _ = bot.DeleteMessage(cb.Message.GetChat().Id, cb.Message.GetMessageId(), nil)
		_, _ = bot.AnswerCallbackQuery(cb.Id, &gotgbot.AnswerCallbackQueryOpts{
			Text: "Проект выбран",
		})
		return nil
	}))
}

func (t *Transport) Run() {
	slog.Info("Bot is running...")
	err := t.updater.StartPolling(t.bot, &ext.PollingOpts{
		DropPendingUpdates: true,
		GetUpdatesOpts: &gotgbot.GetUpdatesOpts{
			Timeout: 10,
		},
	})
	if err != nil {
		panic("failed to start polling: " + err.Error())
	}
	t.updater.Idle()
}

func (t *Transport) handleBotAddedToChat(bot *gotgbot.Bot, ctx *ext.Context) error {
	projects, err := t.service.GetProjects(context.Background())
	if err != nil {
		slog.Error("Failed to get projects", "error", err)
		_, _ = ctx.EffectiveMessage.Reply(bot, "Ошибка при получении проектов", nil)
		return nil
	}

	var keyboard [][]gotgbot.InlineKeyboardButton
	var row []gotgbot.InlineKeyboardButton

	for _, p := range projects {
		btn := gotgbot.InlineKeyboardButton{
			Text:         p.Name,
			CallbackData: p.ID.String(),
		}
		row = append(row, btn)
		if len(row) == 2 {
			keyboard = append(keyboard, row)
			row = nil
		}
	}
	if len(row) > 0 {
		keyboard = append(keyboard, row)
	}

	_, err = ctx.EffectiveMessage.Reply(bot, "Выберете проект:", &gotgbot.SendMessageOpts{
		ReplyMarkup: &gotgbot.InlineKeyboardMarkup{
			InlineKeyboard: keyboard,
		},
	})

	if err != nil {
		slog.Error("Failed to send inline keyboard", "error", err)
	}
	return nil
}

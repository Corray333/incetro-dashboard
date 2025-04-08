package transport

import (
	"context"
	"log"
	"log/slog"
	"os"

	"github.com/corray333/tg-task-parser/internal/entities/project"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
)

type Transport struct {
	service service
	bot     *tgbotapi.BotAPI
}

type service interface {
	CreateTask(ctx context.Context, chatID int64, message string, replyMessage string) error
	GetProjects(ctx context.Context) ([]project.Project, error)
	LinkChatToProject(ctx context.Context, chatID int64, projectID uuid.UUID) error
}

func New(service service) *Transport {
	token := os.Getenv("TASK_PARSER_BOT_TOKEN")

	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("Failed to create bot: ", err)
	}

	bot.Debug = true

	return &Transport{
		service: service,
		bot:     bot,
	}
}

func (t *Transport) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)

	slog.Info("Listening for updates...")

	for update := range updates {
		go t.handleMessage(update)
	}
}

func (t *Transport) handleMessage(update tgbotapi.Update) {
	if update.CallbackQuery != nil {
		t.handleCallbackQuery(update)
		return
	}

	message := update.Message
	if message == nil {
		return
	}

	if message.NewChatMembers != nil {
		for _, member := range message.NewChatMembers {
			if member.ID == t.bot.Self.ID {
				t.handleBotAddedToChat(message.Chat.ID)
				return
			}
		}
		return
	}

	if message.Text == "" {
		return
	}

	mainText := message.Text
	var replyText string
	if message.ReplyToMessage != nil {
		replyText = message.ReplyToMessage.Text
	}

	// Создаем задачу
	err := t.service.CreateTask(context.Background(), message.Chat.ID, mainText, replyText)
	if err != nil {
		slog.Error("Error creating task", "error", err)
		// Отправляем сообщение об ошибке в Telegram
		msg := tgbotapi.NewMessage(message.Chat.ID, "Error creating task")
		_, err := t.bot.Send(msg)
		if err != nil {
			slog.Error("Error sending error message", "error", err)
		}
		return
	}
}

func (t *Transport) handleBotAddedToChat(chatID int64) {
	projects, err := t.service.GetProjects(context.Background())
	if err != nil {
		slog.Error("Error getting projects", "error", err)
		msg := tgbotapi.NewMessage(chatID, "Error getting projects")
		t.bot.Send(msg)
		return
	}

	var keyboard [][]tgbotapi.InlineKeyboardButton
	var row []tgbotapi.InlineKeyboardButton

	for i, project := range projects {
		button := tgbotapi.NewInlineKeyboardButtonData(project.Name, project.ID.String())
		row = append(row, button)

		// Add a row every 2 buttons or at the end
		if len(row) == 2 || i == len(projects)-1 {
			keyboard = append(keyboard, row)
			row = nil // Reset row
		}
	}

	msg := tgbotapi.NewMessage(chatID, "Выберете проект:")
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)

	_, err = t.bot.Send(msg)
	if err != nil {
		slog.Error("Error sending message with inline keyboard", "error", err)
	}
}

func (t *Transport) handleCallbackQuery(update tgbotapi.Update) {
	callbackQuery := update.CallbackQuery
	if callbackQuery == nil {
		return
	}

	projectIDStr := callbackQuery.Data
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		slog.Error("Error parsing project ID", "error", err)
		return
	}

	err = t.service.LinkChatToProject(context.Background(), callbackQuery.Message.Chat.ID, projectID)
	if err != nil {
		slog.Error("Error setting project in chat", "error", err)
		return
	}

	// Delete message and send callback answer
	msg := tgbotapi.NewDeleteMessage(callbackQuery.Message.Chat.ID, callbackQuery.Message.MessageID)
	_, err = t.bot.Request(msg)
	if err != nil {
		slog.Error("Error deleting message", "error", err)
		return
	}

	callback := tgbotapi.NewCallback(callbackQuery.ID, "Проект выбран")
	_, err = t.bot.Request(callback)
	if err != nil {
		slog.Error("Error sending callback", "error", err)
	}
}

// @title Task Tracker API
// @version 1.0
// @description API for task tracking using notion
// @BasePath /tracker

package transport

import (
	"context"
	"log"
	"log/slog"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Transport struct {
	service service
	bot     *tgbotapi.BotAPI
}

type service interface {
	CreateTask(ctx context.Context, message string, replyMessage string) error
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
		go t.handleMessage(update.Message)
	}
}

func (t *Transport) handleMessage(message *tgbotapi.Message) {
	if message == nil {
		return
	}

	// Получаем текст основного сообщения и текста-реплая (если есть)
	mainText := message.Text
	var replyText string
	if message.ReplyToMessage != nil {
		replyText = message.ReplyToMessage.Text
	}

	// Создаем задачу
	err := t.service.CreateTask(context.Background(), mainText, replyText)
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

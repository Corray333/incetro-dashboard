// @title Task Tracker API
// @version 1.0
// @description API for task tracking using notion
// @BasePath /tracker

package transport

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Transport struct {
	service service
	bot     *tgbotapi.BotAPI
}

type service interface {
}

func New(service service) *Transport {
	token := os.Getenv("BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatal("Failed to create bot: ", err)
	}

	bot.Debug = true

	return &Transport{
		service: service,
	}
}

func (t *Transport) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := t.bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		// Получаем текст основного сообщения и текста-реплая (если есть)
		// mainText := update.Message.Text
		// var replyText string
		// if update.Message.ReplyToMessage != nil {
		// 	replyText = update.Message.ReplyToMessage.Text
		// }

		// // Пример ответа в чат
		// msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Задача сохранена:\n%s", task.Text))
		// t.bot.Send(msg)
	}
}

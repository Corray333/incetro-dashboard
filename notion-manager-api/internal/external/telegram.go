package external

import (
	"context"
	"fmt"
	"strings"

	"github.com/Corray333/employee_dashboard/internal/entities"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var Subscribers = []int64{737415136, 377742748, 56218566}

func (e *External) SendNotification(rows []entities.Row) error {
	msgText := fmt.Sprintf("Ошибки в записях пользователя %s:\n", rows[0].Employee)

	for i, row := range rows {
		msgText += fmt.Sprintf("%d. <a href=\"notion.so/%s\">%s</a>\n", i+1, strings.ReplaceAll(row.ID, "-", ""), row.Description)
	}

	for _, subscriber := range Subscribers {
		msg := tgbotapi.NewMessage(subscriber, msgText)
		msg.ParseMode = tgbotapi.ModeHTML
		if _, err := e.tg.GetBot().Send(msg); err != nil {
			return err
		}
	}

	return nil
}

func (e *External) SendSalaryNotification(ctx context.Context, employeeID int64) error {
	msg := tgbotapi.NewMessage(employeeID, "Заработная плата отправлена")
	if _, err := e.tg.GetBot().Send(msg); err != nil {
		return err
	}

	return nil
}

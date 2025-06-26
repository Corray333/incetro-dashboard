package tg

import (
	"context"
	"log/slog"

	"github.com/Corray333/employee_dashboard/internal/domains/weekday/entities/weekday"
	"github.com/Corray333/employee_dashboard/internal/telegram"
)

type WeekdayPostgresRepository struct {
	*telegram.TelegramClient
}

func NewWeekdayTelegramRepository(client *telegram.TelegramClient) *WeekdayPostgresRepository {
	return &WeekdayPostgresRepository{client}
}

var managerIDs = []int64{
	377742748,
	373352303,
	56218566,
	795836353,
}

func (r *WeekdayPostgresRepository) SendWeekendNotification(ctx context.Context, weekday *weekday.Weekday) error {
	msg := weekday.GetNotifyMsg()
	sentMessages := map[int64]int64{}
	for _, id := range managerIDs {
		msg, err := r.GetBot().SendMessage(id, msg, nil)
		if err != nil {
			slog.Error("Error sending message", "error", err)
			// for _, msgID := range sentMessages {
			// 	if _, err := r.GetBot().DeleteMessage(id, msgID, nil); err != nil {
			// 		slog.Error("Error deleting message", "error", err)
			// 	}
			// }
			// return err
		}
		sentMessages[id] = msg.MessageId
	}
	return nil
}

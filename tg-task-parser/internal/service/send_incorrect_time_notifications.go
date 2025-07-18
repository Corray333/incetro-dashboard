package service

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

type incorrectTimeNotificationSender interface {
	SendIncorrectTimeNotifications(ctx context.Context) error
}

type employeesWithIncorrectTimeGetter interface {
	GetEmployeesWithIncorrectTimeEntries(ctx context.Context) ([]uuid.UUID, error)
}

type tgMessageSender interface {
	SendMessage(ctx context.Context, tgID int64, text string) error
}

func (s *Service) SendIncorrectTimeNotifications(ctx context.Context) error {
	// Получаем список сотрудников с ошибочными списаниями
	employeeIDs, err := s.notionRepo.GetEmployeesWithIncorrectTimeEntries(ctx)
	if err != nil {
		slog.Error("Failed to get employees with incorrect time entries", "error", err)
		return err
	}

	if len(employeeIDs) == 0 {
		slog.Info("No employees with incorrect time entries found")
		return nil
	}

	// messageText := "У тебя есть ошибочные списания. Перейди по [ссылке](https://www.notion.so/incetro/22f27b05040480fa8c32e74eabf19777?v=22f27b05040480f39bd0000c39afede1) и исправь, пожалуйста."
	messageText := "Я нашел у тебя в Notion ошибочные списания\n\nПерейди пжлст по [ссылочке](https://notion.so/incetro/22f27b05040480fa8c32e74eabf19777) чтобы поправить их\n\nОни сгруппированы по типу ошибки, поэтому тебе будет легко понять, что поменять, чтобы записи стали корректными"

	// Отправляем уведомления каждому сотруднику
	for _, employeeID := range employeeIDs {
		// Получаем tg_id сотрудника
		tgID, err := s.repository.GetEmployeeTgIDByID(ctx, employeeID)
		if err != nil {
			slog.Error("Failed to get employee tg_id", "employee_id", employeeID, "error", err)
			continue // Пропускаем этого сотрудника и продолжаем с остальными
		}

		// Отправляем сообщение
		if err := s.tgRepo.SendMessage(ctx, tgID, messageText); err != nil {
			slog.Error("Failed to send message to employee", "employee_id", employeeID, "tg_id", tgID, "error", err)
			continue // Пропускаем этого сотрудника и продолжаем с остальными
		}

		slog.Info("Successfully sent notification", "employee_id", employeeID, "tg_id", tgID)
	}

	return nil
}

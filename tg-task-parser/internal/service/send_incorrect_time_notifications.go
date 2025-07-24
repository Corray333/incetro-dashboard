package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

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

var managerTelegramIDs = []int{
	795836353,
	377742748,
	56218566,
	373352303,
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

	// Подсчитываем количество ошибочных записей для каждого сотрудника
	errorCounts := make(map[uuid.UUID]int)
	for _, id := range employeeIDs {
		errorCounts[id]++
	}

	// Обеспечиваем уникальность ID сотрудников
	uniqueEmployeeIDs := make([]uuid.UUID, 0, len(errorCounts))
	for id := range errorCounts {
		uniqueEmployeeIDs = append(uniqueEmployeeIDs, id)
	}

	slog.Info("Processing unique employees", "total_entries", len(employeeIDs), "unique_employees", len(uniqueEmployeeIDs))

	messageText := "Я нашел у тебя в Notion ошибочные списания\n\nПерейди пжлст по [ссылочке](https://notion.so/incetro/22f27b05040480fa8c32e74eabf19777) чтобы поправить их\n\nОни сгруппированы по типу ошибки, поэтому тебе будет легко понять, что поменять, чтобы записи стали корректными"

	// Собираем информацию о сотрудниках для статистики менеджерам
	var employeeStatsList []employeeStats

	// Отправляем уведомления каждому уникальному сотруднику
	for _, employeeID := range uniqueEmployeeIDs {
		// Получаем информацию о сотруднике
		empl, err := s.repository.GetEmployeeByProfileID(ctx, employeeID)
		if err != nil {
			slog.Error("Failed to get employee info", "employee_id", employeeID, "error", err)
			continue // Пропускаем этого сотрудника и продолжаем с остальными
		}

		// Отправляем сообщение сотруднику
		if err := s.tgRepo.SendMessage(ctx, empl.TgID, messageText); err != nil {
			slog.Error("Failed to send message to employee", "employee_id", employeeID, "tg_id", empl.TgID, "error", err)
			continue // Пропускаем этого сотрудника и продолжаем с остальными
		}

		// Добавляем в статистику для менеджеров только после успешной отправки
		employeeStatsList = append(employeeStatsList, employeeStats{
			Name:       empl.FIO,
			TgUsername: empl.TgUsername,
			ErrorCount: errorCounts[employeeID],
		})

		slog.Info("Successfully sent notification", "employee_id", employeeID, "tg_id", empl.TgID)
	}

	// Отправляем статистику менеджерам
	if len(employeeStatsList) > 0 {
		if err := s.sendManagerNotifications(ctx, employeeStatsList); err != nil {
			slog.Error("Failed to send manager notifications", "error", err)
			// Не возвращаем ошибку, так как основная задача (уведомление сотрудников) выполнена
		}
	}

	return nil
}

// sendManagerNotifications отправляет статистику об ошибочных списаниях менеджерам
func (s *Service) sendManagerNotifications(ctx context.Context, employeeStatsList []employeeStats) error {
	// Формируем сообщение для менеджеров
	var messageBuilder strings.Builder
	messageBuilder.WriteString("Сегодня рассылка об ошибочных списаниях была отправлена:\n\n")

	for i, stats := range employeeStatsList {
		username := stats.TgUsername
		if username == "" {
			username = "не указан"
		} else {
			username = "@" + username
		}

		errorText := "ошибочное списание"
		if stats.ErrorCount > 1 {
			if stats.ErrorCount < 5 {
				errorText = "ошибочных списания"
			} else {
				errorText = "ошибочных списаний"
			}
		}

		messageBuilder.WriteString(fmt.Sprintf("%d. %s (тг: %s) — %d %s\n",
			i+1, stats.Name, username, stats.ErrorCount, errorText))
	}

	managerMessage := messageBuilder.String()

	// Отправляем сообщение каждому менеджеру
	for _, managerID := range managerTelegramIDs {
		if err := s.tgRepo.SendMessage(ctx, int64(managerID), managerMessage); err != nil {
			slog.Error("Failed to send message to manager", "manager_id", managerID, "error", err)
			continue // Продолжаем отправку остальным менеджерам
		}

		slog.Info("Successfully sent manager notification", "manager_id", managerID)
	}

	return nil
}

// employeeStats содержит статистику по сотруднику для отправки менеджерам
type employeeStats struct {
	Name       string
	TgUsername string
	ErrorCount int
}

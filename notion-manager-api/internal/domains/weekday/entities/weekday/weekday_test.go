package weekday

import (
	"testing"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/employee/entities/employee"
	"github.com/google/uuid"
)

func TestGetNotifyMsg(t *testing.T) {
	baseTime := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)
	updatedTime := baseTime.Add(10 * time.Minute) // More than 5 minutes difference
	
	tests := []struct {
		name      string
		category  Category
		start     time.Time
		end       time.Time
		reason    string
		createdAt time.Time
		updatedAt time.Time
		employee  employee.Employee
		expects   string
	}{
		{
			name:      "single-day vacation",
			category:  Category("отпуск"),
			start:     time.Date(2025, time.March, 10, 0, 0, 0, 0, time.UTC),
			end:       time.Time{},
			createdAt: baseTime,
			updatedAt: baseTime,
			employee:  employee.Employee{Username: "Mark"},
			expects:   "Mark берёт отпуск на 10 марта",
		},
		{
			name:      "multi-day short vacation",
			category:  Category("отпуск"),
			start:     time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			end:       time.Date(2025, time.January, 3, 0, 0, 0, 0, time.UTC),
			createdAt: baseTime,
			updatedAt: baseTime,
			employee:  employee.Employee{Username: "Mark"},
			expects:   "Mark берёт отпуск с 1 января по 3 января",
		},
		{
			name:      "multi-day long vacation with count",
			category:  Category("отпуск"),
			start:     time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			end:       time.Date(2025, time.January, 5, 0, 0, 0, 0, time.UTC),
			createdAt: baseTime,
			updatedAt: baseTime,
			employee:  employee.Employee{Username: "Mark"},
			expects:   "Mark берёт отпуск с 1 января по 5 января (5 дней)",
		},
		{
			name:      "single-day force majeure",
			category:  CategoryForce,
			start:     time.Date(2025, time.June, 15, 0, 0, 0, 0, time.UTC),
			end:       time.Date(2025, time.June, 15, 0, 0, 0, 0, time.UTC),
			createdAt: baseTime,
			updatedAt: baseTime,
			employee:  employee.Employee{Username: "olga"},
			expects:   "olga форс-мажор — будет отсутствовать на 15 июня",
		},
		{
			name:      "multi-day force majeure with count",
			category:  CategoryForce,
			start:     time.Date(2025, time.June, 1, 0, 0, 0, 0, time.UTC),
			end:       time.Date(2025, time.June, 6, 0, 0, 0, 0, time.UTC), // 6 days
			createdAt: baseTime,
			updatedAt: baseTime,
			employee:  employee.Employee{Username: "olga"},
			expects:   "olga форс-мажор — будет отсутствовать с 1 июня по 6 июня (6 дней)",
		},
		{
			name:      "updated vacation with reason",
			category:  Category("отпуск"),
			start:     time.Date(2025, time.March, 10, 0, 0, 0, 0, time.UTC),
			end:       time.Time{},
			reason:    "семейные обстоятельства",
			createdAt: baseTime,
			updatedAt: updatedTime,
			employee:  employee.Employee{Username: "Mark"},
			expects:   "Mark обновляет отпуск на 10 марта (семейные обстоятельства)",
		},
		{
			name:      "updated long vacation with reason and days count",
			category:  Category("отпуск"),
			start:     time.Date(2025, time.January, 1, 0, 0, 0, 0, time.UTC),
			end:       time.Date(2025, time.January, 5, 0, 0, 0, 0, time.UTC),
			reason:    "продление отпуска",
			createdAt: baseTime,
			updatedAt: updatedTime,
			employee:  employee.Employee{Username: "Mark"},
			expects:   "Mark обновляет отпуск с 1 января по 5 января (5 дней, продление отпуска)",
		},
		{
			name:      "updated force majeure",
			category:  CategoryForce,
			start:     time.Date(2025, time.June, 15, 0, 0, 0, 0, time.UTC),
			end:       time.Date(2025, time.June, 15, 0, 0, 0, 0, time.UTC),
			reason:    "болезнь",
			createdAt: baseTime,
			updatedAt: updatedTime,
			employee:  employee.Employee{Username: "olga"},
			expects:   "olga обновляет форс-мажор — будет отсутствовать на 15 июня (болезнь)",
		},
		{
			name:      "vacation with reason only",
			category:  Category("отпуск"),
			start:     time.Date(2025, time.March, 10, 0, 0, 0, 0, time.UTC),
			end:       time.Time{},
			reason:    "личные дела",
			createdAt: baseTime,
			updatedAt: baseTime,
			employee:  employee.Employee{Username: "Mark"},
			expects:   "Mark берёт отпуск на 10 марта (личные дела)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := Weekday{
				ID:          uuid.New(),
				Category:    tc.category,
				PeriodStart: tc.start,
				PeriodEnd:   tc.end,
				Reason:      tc.reason,
				CreatedAt:   tc.createdAt,
				UpdatedAt:   tc.updatedAt,
				Notified:    false,
				Employee:    tc.employee,
			}
			got := w.GetNotifyMsg()
			if got != tc.expects {
				t.Errorf("expected '%s', got '%s'", tc.expects, got)
			}
		})
	}
}

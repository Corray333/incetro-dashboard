package weekday

import (
	"fmt"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/employee/entities/employee"
	"github.com/google/uuid"
)

type Weekday struct {
	ID          uuid.UUID `json:"id"`
	Category    Category  `json:"category"`
	PeriodStart time.Time `json:"period_start"`
	PeriodEnd   time.Time `json:"period_end"`
	Reason      string    `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Notified    bool      `json:"notified"`

	Employee employee.Employee `json:"employee"`
}

var months = map[time.Month]string{
	time.January:   "января",
	time.February:  "февраля",
	time.March:     "марта",
	time.April:     "апреля",
	time.May:       "мая",
	time.June:      "июня",
	time.July:      "июля",
	time.August:    "августа",
	time.September: "сентября",
	time.October:   "октября",
	time.November:  "ноября",
	time.December:  "декабря",
}

func formatDate(d time.Time) string {
	return fmt.Sprintf("%d %s", d.Day(), months[d.Month()])
}

func (w Weekday) GetNotifyMsg() string {
	username := w.Employee.Username
	periodPhrase := w.getPeriodPhrase()
	categoryPhrase := w.getCategoryPhrase()

	msg := fmt.Sprintf("%s %s %s", username, categoryPhrase, periodPhrase)
	return w.appendDaysCountIfNeeded(msg)
}

func (w Weekday) getPeriodPhrase() string {
	start := formatDate(w.PeriodStart)
	if w.PeriodEnd.IsZero() || w.PeriodStart.Equal(w.PeriodEnd) {
		return "на " + start
	}
	return fmt.Sprintf("с %s по %s", start, formatDate(w.PeriodEnd))
}

func (w Weekday) getCategoryPhrase() string {
	if w.Category == CategoryForce {
		return "форс-мажор — будет отсутствовать"
	}
	return fmt.Sprintf("берёт %s", w.Category)
}

func (w Weekday) appendDaysCountIfNeeded(msg string) string {
	var days int
	if w.PeriodEnd.IsZero() {
		days = 1
	} else {
		days = int(w.PeriodEnd.Sub(w.PeriodStart).Hours()/24) + 1
	}
	if days >= 5 {
		return fmt.Sprintf("%s (%d дней)", msg, days)
	}
	return msg
}

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
	isSingleDay := false
	if w.PeriodEnd.IsZero() {
		isSingleDay = true
	} else {
		isSingleDay = w.PeriodStart.Equal(w.PeriodEnd)
	}

	if w.Category == CategoryForce {
		if isSingleDay {
			return fmt.Sprintf("У %s форс-мажор — будет отсутствовать %s", w.Employee.Username, formatDate(w.PeriodStart))
		} else {
			return fmt.Sprintf("У %s форс-мажор — будет отсутствовать с %s по %s", w.Employee.Username, formatDate(w.PeriodStart), formatDate(w.PeriodEnd))
		}
	}

	if isSingleDay {
		return fmt.Sprintf("%s берёт %s на %s", w.Employee.Username, w.Category, formatDate(w.PeriodStart))
	} else {
		return fmt.Sprintf("%s берёт %s с %s по %s", w.Employee.Username, w.Category, formatDate(w.PeriodStart), formatDate(w.PeriodEnd))
	}
}

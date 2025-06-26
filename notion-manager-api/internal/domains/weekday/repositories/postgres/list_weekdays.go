package postgres

import (
	"context"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/employee/entities/employee"
	"github.com/Corray333/employee_dashboard/internal/domains/weekday/entities/weekday"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type weekdayDB struct {
	ID          uuid.UUID `db:"weekday_id"`
	EmployeeID  uuid.UUID `db:"employee_id"`
	Category    string    `db:"category"`
	PeriodStart time.Time `db:"start_time"`
	PeriodEnd   time.Time `db:"end_time"`
	Reason      string    `db:"reason"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
	Notified    bool      `db:"notified"`
}

func weekdayDBFromEntity(w *weekday.Weekday) *weekdayDB {
	return &weekdayDB{
		ID:          w.ID,
		EmployeeID:  w.Employee.ID,
		Category:    string(w.Category),
		PeriodStart: w.PeriodStart,
		PeriodEnd:   w.PeriodEnd,
		Reason:      w.Reason,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
		Notified:    w.Notified,
	}
}

func (w *weekdayDB) ToEntity() *weekday.Weekday {
	return &weekday.Weekday{
		ID:          w.ID,
		Category:    weekday.Category(w.Category),
		PeriodStart: w.PeriodStart,
		PeriodEnd:   w.PeriodEnd,
		Reason:      w.Reason,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
		Notified:    w.Notified,

		Employee: employee.Employee{
			ID: w.EmployeeID,
		},
	}
}

func (r *WeekdayPostgresRepository) ListWeekdays(ctx context.Context, filter *weekday.Filter) ([]weekday.Weekday, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query := psql.
		Select("*").
		From("weekdays")

	if filter.ID != uuid.Nil {
		query = query.Where(squirrel.Eq{"weekday_id": filter.ID})
	}

	if !filter.UpdatedAtFrom.IsZero() {
		query = query.Where(squirrel.GtOrEq{"updated_at": filter.UpdatedAtFrom})
	}

	if !filter.UpdatedAtTo.IsZero() {
		query = query.Where(squirrel.LtOrEq{"updated_at": filter.UpdatedAtTo})
	}

	if filter.Notified != nil {
		query = query.Where(squirrel.Eq{"notified": *filter.Notified})
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var weekdaysDB []weekdayDB
	if err := r.DB().Select(&weekdaysDB, sqlQuery, args...); err != nil {
		return nil, err
	}

	weekdays := make([]weekday.Weekday, 0, len(weekdaysDB))
	for _, w := range weekdaysDB {
		weekdays = append(weekdays, *w.ToEntity())
		employees, err := r.employeeGetter.ListEmployees(ctx, &employee.Filter{
			ProfileID: w.EmployeeID,
		})
		if err != nil {
			return nil, err
		}

		if len(employees) > 0 {
			weekdays[len(weekdays)-1].Employee = employees[0]
		}
	}

	return weekdays, nil

}

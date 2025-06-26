package postgres

import (
	"context"

	"github.com/Corray333/employee_dashboard/internal/domains/employee/entities/employee"
	"github.com/Corray333/employee_dashboard/internal/postgres"
)

type employeeGetter interface {
	ListEmployees(ctx context.Context, filter *employee.Filter) ([]employee.Employee, error)
}

type WeekdayPostgresRepository struct {
	*postgres.PostgresClient
	employeeGetter employeeGetter
}

func NewWeekdayPostgresRepository(client *postgres.PostgresClient, employeeGetter employeeGetter) *WeekdayPostgresRepository {
	return &WeekdayPostgresRepository{
		PostgresClient: client,
		employeeGetter: employeeGetter,
	}
}

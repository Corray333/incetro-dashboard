package postgres

import (
	"context"

	"github.com/Corray333/employee_dashboard/internal/domains/employee/entities/employee"
	"github.com/Corray333/employee_dashboard/internal/postgres"
)

type employeeGetter interface {
	GetEmployee(ctx context.Context, filter *employee.Filter) ([]employee.Employee, error)
}

type EmployeePostgresRepository struct {
	*postgres.PostgresClient
}

func NewEmployeePostgresRepository(client *postgres.PostgresClient) *EmployeePostgresRepository {
	return &EmployeePostgresRepository{
		PostgresClient: client,
	}
}

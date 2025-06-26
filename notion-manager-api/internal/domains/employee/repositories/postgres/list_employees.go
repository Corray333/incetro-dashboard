package postgres

import (
	"context"

	"github.com/Corray333/employee_dashboard/internal/domains/employee/entities/employee"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

type employeeDB struct {
	ID       uuid.UUID `db:"employee_id"`
	Username string    `db:"username"`
}

func employeeDBFromEntity(e *employee.Employee) *employeeDB {
	return &employeeDB{
		ID:       e.ID,
		Username: e.Username,
	}
}

func (e *employeeDB) ToEntity() *employee.Employee {
	return &employee.Employee{
		ID:       e.ID,
		Username: e.Username,
	}
}

func (r *EmployeePostgresRepository) ListEmployees(ctx context.Context, filter *employee.Filter) ([]employee.Employee, error) {
	employees := []employee.Employee{}
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query := psql.Select("employee_id", "username").From("employees")

	if filter.ID != uuid.Nil {
		query = query.Where(squirrel.Eq{"employee_id": filter.ID})
	}

	if filter.ProfileID != uuid.Nil {
		query = query.Where(squirrel.Eq{"profile_id": filter.ProfileID})
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		return nil, err
	}

	var employeesDB []employeeDB
	if err := r.DB().Select(&employeesDB, sqlQuery, args...); err != nil {
		return nil, err
	}

	for _, e := range employeesDB {
		employees = append(employees, *e.ToEntity())
	}

	return employees, nil

}

package service

import (
	"context"

	"github.com/Corray333/employee_dashboard/internal/domains/employee/entities/employee"
)

type employeeLister interface {
	ListEmployees(ctx context.Context, filter *employee.Filter) ([]employee.Employee, error)
}

func (s *EmployeeService) ListEmployees(ctx context.Context, filter *employee.Filter) ([]employee.Employee, error) {
	return s.employeeLister.ListEmployees(ctx, filter)
}

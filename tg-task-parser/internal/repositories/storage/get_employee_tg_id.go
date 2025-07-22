package storage

import (
	"context"
	"log/slog"

	"github.com/corray333/tg-task-parser/internal/entities/employee"
	"github.com/google/uuid"
)

type employeeDB struct {
	ID         uuid.UUID `db:"employee_id"`
	TgID       int64     `db:"tg_id"`
	TgUsername string    `db:"tg_username"`
	FIO        string    `db:"fio"`
}

func (e *employeeDB) ToEmployee() *employee.Employee {
	return &employee.Employee{
		ID:         e.ID,
		TgID:       e.TgID,
		TgUsername: e.TgUsername,
		FIO:        e.FIO,
	}
}

func (r *PostgresRepository) GetEmployeeByProfileID(ctx context.Context, profileID uuid.UUID) (*employee.Employee, error) {
	empl := &employeeDB{}
	if err := r.db.Get(&empl, "SELECT employee_id, tg_id, tg_username, fio FROM employees WHERE profile_id = $1", profileID); err != nil {
		slog.Error("Error while getting employee tg_id by ID", "profile_id", profileID, "error", err)
		return nil, err
	}
	return empl.ToEmployee(), nil
}

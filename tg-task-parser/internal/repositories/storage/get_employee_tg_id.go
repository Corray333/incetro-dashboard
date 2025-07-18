package storage

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
)

func (r *PostgresRepository) GetEmployeeTgIDByID(ctx context.Context, employeeID uuid.UUID) (int64, error) {
	var tgID int64
	if err := r.db.Get(&tgID, "SELECT tg_id FROM employees WHERE employee_id = $1", employeeID); err != nil {
		slog.Error("Error while getting employee tg_id by ID", "employee_id", employeeID, "error", err)
		return 0, err
	}
	return tgID, nil
}

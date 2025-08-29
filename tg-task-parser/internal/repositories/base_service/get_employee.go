package base_service

import (
	"context"
	"log/slog"
)

func (r *BaseService) GetEmployeeByTgUsername(ctx context.Context, username string) (string, error) {
	employeeID := ""
	err := r.db.QueryRow("SELECT employee_id FROM employees WHERE tg_username = $1", username).Scan(&employeeID)
	if err != nil {
		slog.Error("Notion error while getting employee ID", "error", err, "username", username)
		return "", err
	}
	return employeeID, nil
}

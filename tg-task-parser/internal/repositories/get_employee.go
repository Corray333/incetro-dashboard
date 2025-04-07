package repositories

import "context"

func (r *Repository) GetEmployeeByTgUsername(ctx context.Context, username string) (string, error) {
	employeeID := ""
	err := r.db.QueryRowContext(ctx, "SELECT employee_id FROM employees WHERE tg_username = $1", username).Scan(&employeeID)
	if err != nil {
		return "", err
	}
	return employeeID, nil
}

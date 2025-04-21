package base_service

import "context"

func (r *BaseService) GetEmployeeByTgUsername(ctx context.Context, username string) (string, error) {
	employeeID := ""
	err := r.db.QueryRow("SELECT employee_id FROM employees WHERE tg_username = $1", username).Scan(&employeeID)
	if err != nil {
		return "", err
	}
	return employeeID, nil
}

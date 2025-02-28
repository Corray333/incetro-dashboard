package repositories

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/Corray333/employee_dashboard/internal/entities"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

type Storage struct {
	db *sqlx.DB

	dashboardUsers map[string]entities.Employee
}

func New() *Storage {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", os.Getenv("POSTGRES_HOST"), os.Getenv("POSTGRES_PORT"), os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), os.Getenv("POSTGRES_DB_NAME"))
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	if err := db.Ping(); err != nil {
		panic(err)
	}

	return &Storage{
		db:             db,
		dashboardUsers: map[string]entities.Employee{},
	}
}

func (s *Storage) GetEmployees() (employees []entities.Employee, err error) {
	if err := s.db.Select(&employees, "SELECT * FROM employees"); err != nil {
		slog.Error("error getting employees: " + err.Error())
		return nil, err
	}

	return employees, nil
}

func (s *Storage) GetProjects(userID string) (projects []entities.Project, err error) {
	if err := s.db.Select(&projects, "SELECT projects.*, username as manager FROM projects JOIN employees ON projects.manager_id = employees.profile_id"); err != nil {
		slog.Error("error getting projects: " + err.Error())
		return nil, err
	}

	return projects, nil
}

func (s *Storage) GetActiveTasks(userID string, projectID string) (tasks []entities.Task, err error) {
	statuses := []string{"Формируется", "Можно делать", "На паузе", "Ожидание", "В работе", "Надо обсудить", "Код-ревью", "Внутренняя проверка"}
	query := `
        SELECT * FROM tasks 
        WHERE project_id = $1 
        AND employee_id = $2 
        AND status = ANY($3)
    `
	args := []interface{}{projectID, userID, pq.Array(statuses)}

	if err := s.db.Select(&tasks, query, args...); err != nil {
		slog.Error("error getting tasks: " + err.Error())
		return nil, err
	}

	return tasks, nil
}

func (s *Storage) GetQuarterTasks(quarter int) (tasks []entities.Task, err error) {

	if err := s.db.Select(&tasks, `
        SELECT tasks.* FROM task_tag NATURAL JOIN tasks 
        WHERE tag = $1
    `, "Q"+strconv.Itoa(quarter)); err != nil {
		slog.Error("error getting tasks: " + err.Error())
		return nil, err
	}

	return tasks, nil
}

func (s *Storage) GetUserRole(username string, userID int64) entities.DashboardRole {
	if user, ok := s.dashboardUsers[username]; ok {
		if user.TelegramID == 0 {
			user.TelegramID = userID
			s.dashboardUsers[username] = user
			if err := s.SetEmployeeTelegramID(username, userID); err != nil {
				slog.Error("Error setting tg id of employee", slog.String("error", err.Error()))
				return entities.DashboardRoleUnknown
			}
		}
		return user.Role
	}

	return entities.DashboardRoleUnknown
}

func (s *Storage) SetEmployeeTelegramID(username string, telegramID int64) error {
	if _, err := s.db.Exec("UPDATE employees SET tg_id = $1 WHERE tg_username = $2", telegramID, username); err != nil {
		slog.Error("Error settimg tg id of employee", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (s *Storage) SetEmployees(employees []entities.Employee) error {
	tx, err := s.db.Beginx()
	if err != nil {
		slog.Error("error starting transaction: " + err.Error())
		return err
	}
	defer tx.Rollback()

	for _, employee := range employees {

		var tgID int64 // Use pointer to handle NULL values

		// fmt.Println()
		// fmt.Println(employee)
		// fmt.Println()

		query := `INSERT INTO employees (employee_id, username, email, icon, profile_id, tg_username, geo, expertise, direction, status, phone) 
				  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) 
				  ON CONFLICT (employee_id) 
				  DO UPDATE SET username = $2, email = $3, icon = $4, profile_id = $5, tg_username = $6, geo = $7, expertise = $8, direction = $9, status = $10, phone = $11
				  RETURNING tg_id`

		err := tx.QueryRow(query, employee.ID, employee.Username, employee.Email, employee.Icon, employee.ProfileID, employee.Telegram, employee.Geo, employee.Expertise, employee.Direction, employee.Status, employee.Phone).Scan(&tgID)
		if err != nil {
			slog.Error("error setting employees: " + err.Error())
			return err
		}

		employee.Role = entities.DashboardRoleAdmin
		if tgID != 0 {
			employee.TelegramID = tgID
		}

		s.dashboardUsers[employee.Telegram] = employee

		for _, flag := range employee.NotificationFlags {
			if _, err := tx.Exec("INSERT INTO employee_notification_flag (employee_id, flag) VALUES ($1, $2) ON CONFLICT (employee_id, flag) DO UPDATE SET flag = $2", employee.ID, flag); err != nil {
				slog.Error("Error setting notification flags", slog.String("error", err.Error()))
				return err
			}
		}
	}

	return tx.Commit()
}

func (s *Storage) GetEmployeesByNotificationFlag(ctx context.Context, flag entities.NotificationFlag) (employees []entities.Employee, err error) {
	query := `
		SELECT e.* FROM employees e
		JOIN employee_notification_flag enf ON e.employee_id = enf.employee_id
		WHERE enf.flag = $1
	`
	if err := s.db.Select(&employees, query, flag); err != nil {
		slog.Error("error getting employees by notification flag: " + err.Error())
		return nil, err
	}

	return employees, nil
}

// SetTasks inserts tasks into the postgres database or updates them if they already exist with this uuid
func (s *Storage) SetTasks(tasks []entities.Task) error {
	tx, err := s.db.Beginx()
	if err != nil {
		slog.Error("Error starting transaction", slog.String("error", err.Error()))
		return err
	}
	defer tx.Rollback()

	for _, task := range tasks {
		_, err := tx.Exec("INSERT INTO tasks (task_id, project_id, employee_id, title, status, start_time, end_time) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (task_id) DO UPDATE SET title = $4, status = $5, employee_id = $3, project_id = $2, start_time = $6, end_time = $7", task.ID, task.ProjectID, task.EmployeeID, task.Title, task.Status, task.StartTime, task.EndTime)
		if err != nil {
			slog.Error("Error setting tasks", slog.String("error", err.Error()))
			return err
		}
		for _, tag := range task.Tags {
			if _, err := tx.Exec("INSERT INTO task_tag (task_id, tag) VALUES ($1, $2) ON CONFLICT (task_id, tag) DO UPDATE SET tag = $2", task.ID, tag); err != nil {
				slog.Error("Error setting task tags", slog.String("error", err.Error()))
				return err
			}
		}
	}

	return tx.Commit()
}

func (s *Storage) GetTasksOfEmployee(employeeUsername string, period_start, period_end int64) ([]entities.Task, error) {
	tasks := []entities.Task{}
	query := `
		SELECT tasks.* FROM tasks NATURAL JOIN employees
		WHERE tg_username = $1
		AND (start_time <= $3 AND end_time >= $2)
	`
	if err := s.db.Select(&tasks, query, employeeUsername, period_start, period_end); err != nil {
		slog.Error("error getting tasks of employee: " + err.Error())
		return nil, err
	}
	fmt.Println(tasks)
	return tasks, nil
}

func (s *Storage) SetProjects(projects []entities.Project) error {
	tx, err := s.db.Beginx()
	if err != nil {
		slog.Error("Error starting transaction", slog.String("error", err.Error()))
		return err
	}
	defer tx.Rollback()

	for _, project := range projects {
		_, err := tx.Exec("INSERT INTO projects (project_id, name, icon, icon_type, status, type, manager_id) VALUES ($1, $2, $3, $4, $5, $6, $7) ON CONFLICT (project_id) DO UPDATE SET name = $2, icon = $3, icon_type = $4, status = $5, type = $6, manager_id = $7", project.ID, project.Name, project.Icon, project.IconType, project.Status, project.Type, project.ManagerID)
		if err != nil {
			slog.Error("Error setting projects", slog.String("error", err.Error()))
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) SaveTimeWriteOf(time *entities.TimeMsg) error {
	_, err := s.db.Exec("INSERT INTO time_outbox (task_id, employee_id, duration, description) VALUES ($1, $2, $3, $4)", time.TaskID, time.EmployeeID, time.Duration, time.Description)
	if err != nil {
		slog.Error("Error saving time outbox message", slog.String("error", err.Error()))
		return err
	}

	return nil
}

func (s *Storage) GetTimesMsg() (times []entities.TimeMsg, err error) {
	if err = s.db.Select(&times, "SELECT * FROM time_outbox"); err != nil {
		slog.Error("error getting times messages", slog.String("error", err.Error()))
		return nil, err
	}

	return times, nil
}

func (s *Storage) GetTimes() (times []entities.Time, err error) {
	if err = s.db.Select(&times, "SELECT * FROM times"); err != nil {
		slog.Error("error getting times", slog.String("error", err.Error()))
		return nil, err
	}

	return times, nil
}

func (s *Storage) GetInvalidRows() (times []entities.Row, err error) {
	if err = s.db.Select(&times, "SELECT * FROM invalid_rows"); err != nil {
		slog.Error("error getting invalid rows", slog.String("error", err.Error()))
		return nil, err
	}

	return times, nil
}

func (s *Storage) MarkInvalidRowsAsSent(rows []entities.Row) error {
	tx, err := s.db.Beginx()
	if err != nil {
		slog.Error("error starting transaction", slog.String("error", err.Error()))
		return err
	}
	defer tx.Rollback()

	for _, row := range rows {
		if _, err := tx.Exec("DELETE FROM invalid_rows WHERE id = $1", row.ID); err != nil {
			slog.Error("error deleting invalid row", slog.String("error", err.Error()))
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) SetInvalidRows(rows []entities.Row) error {
	tx, err := s.db.Beginx()
	if err != nil {
		slog.Error("error starting transaction", slog.String("error", err.Error()))
		return err
	}
	defer tx.Rollback()

	for _, row := range rows {
		_, err := tx.Exec("INSERT INTO invalid_rows (id, description, employee, employee_id) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO UPDATE SET description = $2, employee = $3, employee_id = $4", row.ID, row.Description, row.Employee, row.EmployeeID)
		if err != nil {
			slog.Error("error setting invalid rows", slog.String("error", err.Error()))
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) SetTimes(times []entities.Time) error {
	tx, err := s.db.Beginx()
	if err != nil {
		slog.Error("error starting transaction", slog.String("error", err.Error()))
		return err
	}
	defer tx.Rollback()

	for _, time := range times {
		_, err := tx.Exec("INSERT INTO times (time_id, employee, description) VALUES ($1, $2, $3) ON CONFLICT (time_id) DO UPDATE SET  employee = $2, description = $3", time.ID, time.Employee, time.Description)
		if err != nil {
			slog.Error("error setting times: " + err.Error())
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) MarkTimeAsSent(timeID int64) error {
	if _, err := s.db.Exec("DELETE FROM time_outbox WHERE time_id = $1", timeID); err != nil {
		slog.Error("error marking time as sent: " + err.Error())
		return err
	}

	return nil
}

func (s *Storage) GetSystemInfo() (*entities.System, error) {
	system := entities.System{}
	if err := s.db.Get(&system, "SELECT * FROM system"); err != nil {
		slog.Error("error getting system info: " + err.Error())
		return nil, err
	}

	return &system, nil
}

func (s *Storage) SetSystemInfo(system *entities.System) error {
	tx, err := s.db.Beginx()
	if err != nil {
		slog.Error("error starting transaction: " + err.Error())
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE system SET projects_db_last_sync = $1, tasks_db_last_sync = $2, employee_db_last_sync = $3, times_db_last_sync = $4", system.ProjectsDBLastSynced, system.TasksDBLastSynced, system.EmployeeDBLastSynced, system.TimesDBLastSynced)
	if err != nil {
		slog.Error("error updating system info: " + err.Error())
		return err
	}

	return tx.Commit()
}

func (s *Storage) GetEmployeeByID(employeeID string) (employee entities.Employee, err error) {
	if err := s.db.Get(&employee, "SELECT * FROM employees WHERE employee_id = $1", employeeID); err != nil {
		slog.Error("error getting employee by id: " + err.Error())
		return entities.Employee{}, err
	}

	return employee, nil
}

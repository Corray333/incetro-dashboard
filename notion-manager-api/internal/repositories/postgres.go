package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/Corray333/employee_dashboard/internal/entities"
	"github.com/google/uuid"
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
	if err := s.db.Select(&employees, "SELECT employees.*, expertise.name as expertise_name FROM employees NATURAL JOIN expertise ORDER BY unique_id"); err != nil {
		slog.Error("error getting employees: " + err.Error())
		return nil, err
	}

	return employees, nil
}

func (s *Storage) GetProjects(userID string) (projects []entities.Project, err error) {
	query := `
		SELECT 
			projects.*, 
			COALESCE(employees.username, '') as manager,
			COALESCE(clients.name, '') as client
		FROM projects 
		LEFT JOIN employees ON projects.manager_id = employees.profile_id 
		LEFT JOIN clients ON projects.client_id = clients.client_id
		WHERE projects.name != ''
		ORDER BY projects.unique_id ASC
	`
	if err := s.db.Select(&projects, query); err != nil {
		slog.Error("error getting projects: " + err.Error())
		return nil, err
	}

	return projects, nil
}

func (s *Storage) GetProjectsWithHoursSums(ctx context.Context) ([]entities.Project, error) {
	projects := []entities.Project{}
	if err := s.db.Select(&projects, `
		SELECT
			p.project_id,
			COALESCE(SUM(t.estimate) FILTER (
				WHERE t.title NOT IN (
					'Менеджмент ' || p.name,
					'Тестирование ' || p.name
				)
			), 0) AS total_hours,
			COALESCE(MAX(t.task_id::text) FILTER (
				WHERE t.title = 'Менеджмент ' || p.name
			)::text, '') AS management_task_id,
			COALESCE(MAX(t.task_id::text) FILTER (
				WHERE t.title = 'Тестирование ' || p.name
			)::text, '') AS testing_task_id
		FROM
			projects p
		JOIN
			tasks t ON p.project_id = t.project_id
		GROUP BY
			p.project_id, p.name;
	`); err != nil {
		slog.Error("Error while getting projects with hours sums", "error", err)
		return nil, err
	}
	return projects, nil
}

func (s *Storage) GetActiveTasks(userID string, projectID string) (tasks []entities.Task, err error) {
	statuses := []string{"Формируется", "Можно делать", "На паузе", "Ожидание", "В работе", "Надо обсудить", "Код-ревью", "Внутренняя проверка"}
	query := `
        SELECT task_id, title, status, project_id, estimate FROM tasks 
        WHERE project_id = $1 
        AND executor_id = $2 
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
	tasks = []entities.Task{}
	if err := s.db.Select(&tasks, `
        SELECT tasks.task_id, tasks.status FROM task_tag NATURAL JOIN tasks 
        WHERE tag = $1
    `, "Q"+strconv.Itoa(quarter)); err != nil && err != sql.ErrNoRows {
		slog.Error("error getting tasks: " + err.Error())
		return nil, err
	}
	if len(tasks) == 0 {
		return []entities.Task{}, nil
	}

	return tasks, nil
}

func (s *Storage) GetUserRole(username string, userID int64) entities.DashboardRole {
	if user, ok := s.dashboardUsers[username]; ok {
		if user.TelegramID == 0 {
			user.TelegramID = userID
			s.dashboardUsers[username] = user
			if err := s.SetEmployeeTelegramID(username, userID); err != nil {
				slog.Error("Error setting tg id of employee", "error", err)
				return entities.DashboardRoleUnknown
			}
		}
		return user.Role
	}

	return entities.DashboardRoleUnknown
}

func (s *Storage) SetEmployeeTelegramID(username string, telegramID int64) error {
	if _, err := s.db.Exec("UPDATE employees SET tg_id = $1 WHERE tg_username = $2", telegramID, username); err != nil {
		slog.Error("Error settimg tg id of employee", "error", err)
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

		var expertiseID sql.NullString
		if employee.ExpertiseID != "" {
			// Проверяем, что строка является допустимым UUID
			if _, err := uuid.Parse(employee.ExpertiseID); err == nil {
				expertiseID = sql.NullString{String: employee.ExpertiseID, Valid: true}
			} else {
				// Обработка ошибки парсинга UUID
				slog.Error("invalid UUID format for ExpertiseID: " + err.Error())
				return err
			}
		} else {
			expertiseID = sql.NullString{Valid: false}
		}

		query := `INSERT INTO employees (employee_id, username, email, icon, profile_id, tg_username, geo, expertise_id, direction, status, phone, fio, unique_id) 
				  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
				  ON CONFLICT (employee_id) 
				  DO UPDATE SET username = $2, email = $3, icon = $4, profile_id = $5, tg_username = $6, geo = $7, expertise_id = $8, direction = $9, status = $10, phone = $11, fio = $12, unique_id = $13
				  RETURNING tg_id`

		err := tx.QueryRow(query, employee.ID, employee.Username, employee.Email, employee.Icon, employee.ProfileID, employee.Telegram, employee.Geo, expertiseID, employee.Direction, employee.Status, employee.Phone, employee.FIO, employee.UniqueID).Scan(&tgID)
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
				slog.Error("Error setting notification flags", "error", err)
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
		slog.Error("Error starting transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	for _, task := range tasks {
		_, err := tx.Exec("INSERT INTO tasks (task_id, project_id, employee_id, title, status, start_time, end_time, estimate) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) ON CONFLICT (task_id) DO UPDATE SET title = $4, status = $5, employee_id = $3, project_id = $2, start_time = $6, end_time = $7, estimate = $8", task.ID, task.ProjectID, task.ExecutorID, task.Title, task.Status, task.StartTime, task.EndTime, task.Estimate)
		if err != nil {
			slog.Error("Error setting tasks", "error", err)
			return err
		}
		for _, tag := range task.Tags {
			if _, err := tx.Exec("INSERT INTO task_tag (task_id, tag) VALUES ($1, $2) ON CONFLICT (task_id, tag) DO UPDATE SET tag = $2", task.ID, tag); err != nil {
				slog.Error("Error setting task tags", "error", err)
				return err
			}
		}
	}

	return tx.Commit()
}

// type Task struct {
// 	ID         string    `json:"id" db:"task_id" example:"9eb9de5f-2341-44c6-aae8-fc917394092b"`
// 	Title      string    `json:"title" db:"title" example:"Доделать прототип тайм трекера"`
// 	Status     string    `json:"status" db:"status" example:"В работе"`
// 	ProjectID  string    `json:"projectID" db:"project_id" example:"268c4871-39fd-4c78-9681-4d62ae34dcee"`
// 	EmployeeID string    `json:"employeeID" db:"employee_id" example:"353198d1-1a40-4b4b-9841-66e7de4de6ea"`
// 	Employee   string    `json:"employee" example:"Mark"`
// 	Tags       []string  `json:"tags" db:"tags"`
// 	StartTime  time.Time `json:"startTime" db:"start"`
// 	EndTime    time.Time `json:"endTime" db:"end"`
// 	Estimate   float64   `json:"estimate" db:"estimate"`
// }

func (s *Storage) GetTasksOfEmployee(employeeUsername string, period_start, period_end int64, quarter int) ([]entities.Task, error) {
	tasks := []entities.Task{}
	query := `
		SELECT DISTINCT tasks.task_id, tasks.status, tasks.title, tasks.status, tasks.project_id, employees.username as employee, tasks.start, tasks.end, tasks.estimate
		FROM tasks 
		JOIN employees ON tasks.executor_id = employees.employee_id  JOIN task_tag ON tasks.task_id = task_tag.task_id
		WHERE tg_username = $1
		AND (
			(start >= $2 AND start <= $3)
			OR ("end" >= $2 AND "end" <= $3)
		) AND tag = $4
	`
	if err := s.db.Select(&tasks, query, employeeUsername, time.Unix(period_start, 0), time.Unix(period_end, 0), "Q"+strconv.Itoa(quarter)); err != nil && err != sql.ErrNoRows {
		slog.Error("error getting tasks of employee: " + err.Error())
		return nil, err
	}
	if len(tasks) == 0 {
		return []entities.Task{}, nil
	}

	return tasks, nil
}

func (s *Storage) SetProjects(projects []entities.Project) error {
	tx, err := s.db.Beginx()
	if err != nil {
		slog.Error("Error starting transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM projects"); err != nil {
		slog.Error("Error deleting projects", "error", err)
		return err
	}

	for _, project := range projects {
		_, err := tx.Exec("INSERT INTO projects (project_id, name, icon, icon_type, status, type, manager_id, sheets_link, unique_id, client_id) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) ON CONFLICT (project_id) DO UPDATE SET name = $2, icon = $3, icon_type = $4, status = $5, type = $6, manager_id = $7, sheets_link = $8, unique_id = $9, client_id = $10", project.ID, project.Name, project.Icon, project.IconType, project.Status, project.Type, project.ManagerID, project.SheetsLink, project.UniqueID, project.ClientID)
		if err != nil {
			slog.Error("Error setting projects", "error", err)
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) SaveTimeWriteOf(time *entities.TimeMsg) error {
	_, err := s.db.Exec("INSERT INTO time_outbox (task_id, employee_id, duration, description) VALUES ($1, $2, $3, $4)", time.TaskID, time.EmployeeID, time.Duration, time.Description)
	if err != nil {
		slog.Error("Error saving time outbox message", "error", err)
		return err
	}

	return nil
}

func (s *Storage) GetTimesMsg() (times []entities.TimeMsg, err error) {
	if err = s.db.Select(&times, "SELECT * FROM time_outbox"); err != nil {
		slog.Error("error getting times messages", "error", err)
		return nil, err
	}

	return times, nil
}

func (s *Storage) GetTimes() (times []entities.Time, err error) {
	if err = s.db.Select(&times, "SELECT * FROM times"); err != nil {
		slog.Error("error getting times", "error", err)
		return nil, err
	}

	return times, nil
}

func (s *Storage) GetInvalidRows() (times []entities.Row, err error) {
	if err = s.db.Select(&times, "SELECT * FROM invalid_rows"); err != nil {
		slog.Error("error getting invalid rows", "error", err)
		return nil, err
	}

	return times, nil
}

func (s *Storage) MarkInvalidRowsAsSent(rows []entities.Row) error {
	tx, err := s.db.Beginx()
	if err != nil {
		slog.Error("error starting transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	for _, row := range rows {
		if _, err := tx.Exec("DELETE FROM invalid_rows WHERE id = $1", row.ID); err != nil {
			slog.Error("error deleting invalid row", "error", err)
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) SetInvalidRows(rows []entities.Row) error {
	tx, err := s.db.Beginx()
	if err != nil {
		slog.Error("error starting transaction", "error", err)
		return err
	}
	defer tx.Rollback()

	for _, row := range rows {
		_, err := tx.Exec("INSERT INTO invalid_rows (id, description, employee, employee_id) VALUES ($1, $2, $3, $4) ON CONFLICT (id) DO UPDATE SET description = $2, employee = $3, employee_id = $4", row.ID, row.Description, row.Employee, row.EmployeeID)
		if err != nil {
			slog.Error("error setting invalid rows", "error", err)
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) SetTimes(times []entities.Time) error {
	tx, err := s.db.Beginx()
	if err != nil {
		slog.Error("error starting transaction", "error", err)
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

	_, err = tx.Exec("UPDATE system SET projects_db_last_sync = $1, employee_db_last_sync = $2", system.ProjectsDBLastSynced, system.EmployeeDBLastSynced)
	if err != nil {
		slog.Error("error updating system info: " + err.Error())
		return err
	}

	return tx.Commit()
}

func (s *Storage) GetEmployeeByID(employeeID string) (employee entities.Employee, err error) {
	if err := s.db.Get(&employee, "SELECT * FROM employees WHERE employee_id = $1", employeeID); err != nil {
		if err == sql.ErrNoRows {
			return entities.Employee{}, nil
		}
		slog.Error("error getting employee by id: " + err.Error())
		return entities.Employee{}, err
	}

	return employee, nil
}

func (s *Storage) SetExpertises(ctx context.Context, expertises []entities.Expertise) error {
	tx, err := s.db.Beginx()
	if err != nil {
		slog.Error("error starting transaction: " + err.Error())
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM expertise"); err != nil {
		slog.Error("error deleting expertises: " + err.Error())
		return err
	}

	for _, expertise := range expertises {
		_, err := tx.Exec("INSERT INTO expertise (expertise_id, name, direction, description) VALUES ($1, $2, $3, $4) ON CONFLICT (expertise_id) DO UPDATE SET direction = $3, description = $4", expertise.ID, expertise.Name, expertise.Direction, expertise.Description)
		if err != nil {
			slog.Error("error setting expertises: " + err.Error())
			return err
		}
	}

	return tx.Commit()
}

func (s *Storage) GetExpertises() (expertises []entities.Expertise, err error) {
	if err := s.db.Select(&expertises, "SELECT * FROM expertise ORDER BY expertise_id"); err != nil {
		slog.Error("error getting expertises: " + err.Error())
		return nil, err
	}

	return expertises, nil
}

func (s *Storage) GetExtertiseByID(ctx context.Context, id string) (expertise entities.Expertise, err error) {
	if err := s.db.Get(&expertise, "SELECT * FROM expertise WHERE expertise_id = $1", id); err != nil {
		slog.Error("error getting expertise by id: " + err.Error())
		return entities.Expertise{}, err
	}

	return expertise, nil
}

func (s *Storage) DeleteFeedback(ctx context.Context, feedbackID uuid.UUID) error {
	if _, err := s.db.Exec("DELETE FROM feedbacks WHERE feedback_id = $1", feedbackID); err != nil {
		slog.Error("Error deleting feedback", "error", err)
		return err
	}

	return nil
}

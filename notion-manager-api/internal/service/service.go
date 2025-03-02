package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/Corray333/employee_dashboard/internal/entities"
	"github.com/Corray333/employee_dashboard/pkg/mindmap"
	"github.com/go-co-op/gocron"
	"google.golang.org/api/sheets/v4"
)

type repository interface {
	GetEmployees() (employees []entities.Employee, err error)
	GetProjects(userID string) (projects []entities.Project, err error)
	GetActiveTasks(userID string, projectID string) (tasks []entities.Task, err error)
	GetTimesMsg() (times []entities.TimeMsg, err error)

	SetEmployees(employees []entities.Employee) error
	SetTasks(tasks []entities.Task) error
	SetProjects(projects []entities.Project) error
	SaveTimeWriteOf(time *entities.TimeMsg) error

	GetInvalidRows() (times []entities.Row, err error)
	SetInvalidRows(times []entities.Row) error
	MarkInvalidRowsAsSent(times []entities.Row) error

	GetSystemInfo() (*entities.System, error)

	SetSystemInfo(system *entities.System) error
	MarkTimeAsSent(timeID int64) error

	GetEmployeeByID(employeeID string) (employee entities.Employee, err error)

	GetTasksOfEmployee(employee_id string, period_start, period_end int64) ([]entities.Task, error)
	GetQuarterTasks(quarter int) (tasks []entities.Task, err error)
	GetEmployeesByNotificationFlag(ctx context.Context, flag entities.NotificationFlag) (employees []entities.Employee, err error)
	GetUserRole(username string, userID int64) entities.DashboardRole
}

type external interface {
	GetEmployees(lastSynced int64) (employees []entities.Employee, lastUpdate int64, err error)
	GetTasks(timeFilterType string, lastSynced int64, startCursor string, useTitleFilter bool) (tasks []entities.Task, lastUpdate int64, err error)
	GetProjects(lastSynced int64) (projects []entities.Project, lastUpdate int64, err error)
	GetTimes(timeFilterType string, lastSynced int64, startCursor string, useWhatDidFilter bool) (times []entities.Time, lastUpdate int64, err error)

	WriteOfTime(time *entities.TimeMsg) error

	SendNotification(msg []entities.Row) error

	GetNotCorrectPersonTimes() (times []entities.Time, lastUpdate int64, err error)
	SetProfileInTime(timeID, profileID string) error

	NewSheetsClient() (*sheets.Service, error)
	CreateMindmapTasks(projectName string, tasks []mindmap.Task) error
	SendSalaryNotification(ctx context.Context, employeeID int64) error

	UpdateTimeSheet(srv *sheets.Service) error
	UpdateProjectsSheet(srv *sheets.Service, projects []entities.Project) error
	UpdatePeopleSheet(srv *sheets.Service, employees []entities.Employee) error
}

type Service struct {
	repo     repository
	external external
	cron     *gocron.Scheduler
}

func New(repo repository, external external) *Service {
	loc, _ := time.LoadLocation("Europe/Moscow")
	s := gocron.NewScheduler(loc)

	return &Service{
		repo:     repo,
		external: external,
		cron:     s,
	}
}

func (s *Service) Run() {
	go s.StartUpdatingWorker()
	go s.StartOutboxWorker()

	s.cron.Every(1).Day().At("10:00").Do(s.CheckInvalid)
	s.cron.StartBlocking()
}

const CheckAfter = 1727740840

func (s *Service) GetUserRole(username string, userID int64) entities.DashboardRole {
	return s.repo.GetUserRole(username, userID)
}

func (s *Service) CheckInvalid() {
	tasks, _, err := s.external.GetTasks("created_time", CheckAfter, "", true)
	if err != nil {
		slog.Error("error getting tasks: " + err.Error())
	}

	invalid := s.ValidateTasks(tasks)
	if len(invalid) > 0 {
		if err := s.repo.SetInvalidRows(invalid); err != nil {
			slog.Error("error setting invalid rows: " + err.Error())
		}
	}

	fmt.Println("Getting times")
	times, _, err := s.external.GetTimes("created_time", CheckAfter, "", true)
	if err != nil {
		slog.Error("error getting times: " + err.Error())
	}

	invalidTimes := s.ValidateTimes(times)
	invalid = append(invalid, invalidTimes...)

	grouped := s.groupByEmployeeID(invalid)
	for _, rows := range grouped {
		if err := s.external.SendNotification(rows); err != nil {
			slog.Error("error sending notification: " + err.Error())
			continue
		}

		if err := s.repo.MarkInvalidRowsAsSent(rows); err != nil {
			slog.Error("error marking invalid rows as sent: " + err.Error())
			continue
		}
	}

}

func (s *Service) UpdateGoogleSheets() error {
	projects, err := s.repo.GetProjects("")
	if err != nil {
		return err
	}

	srv, err := s.external.NewSheetsClient()
	if err != nil {
		return err
	}

	if err := s.external.UpdateTimeSheet(srv); err != nil {
		return err
	}

	if err := s.external.UpdateProjectsSheet(srv, projects); err != nil {
		return err
	}

	employees, err := s.repo.GetEmployees()
	if err != nil {
		return err
	}
	if err := s.external.UpdatePeopleSheet(srv, employees); err != nil {
		return err
	}

	return nil

}

func (s *Service) GetTasksOfEmployee(employee_username string, period_start, period_end int64) ([]entities.Task, error) {
	return s.repo.GetTasksOfEmployee(employee_username, period_start, period_end)
}

func (s *Service) GetQuarterTasks() ([]entities.Task, error) {
	currentQuarter := (time.Now().Month() + 1) / 3
	return s.repo.GetQuarterTasks(int(currentQuarter))
}

func (s *Service) groupByEmployeeID(rows []entities.Row) map[string][]entities.Row {
	grouped := map[string][]entities.Row{}
	for _, row := range rows {
		if row.Employee == "" && row.EmployeeID != "" {
			employee, err := s.repo.GetEmployeeByID(row.EmployeeID)
			if err == nil {
				row.Employee = employee.Username
			}
		}
		grouped[row.Employee] = append(grouped[row.Employee], row)
	}
	return grouped
}

func (s *Service) StartUpdatingWorker() {
	for {
		_, err := s.Actualize()
		if err != nil {
			slog.Error(err.Error())
		}
		time.Sleep(time.Minute)
	}
}

func (s *Service) StartOutboxWorker() {
	for {
		slog.Info("Reading outbox")
		times, err := s.repo.GetTimesMsg()
		if err != nil {
			slog.Error("error getting times: " + err.Error())
			continue
		}
		for _, time := range times {
			if err := s.external.WriteOfTime(&time); err != nil {
				slog.Error("error sending time to notion: " + err.Error())
				continue
			}

			// TODO: maybe add compensation of notion query
			if err := s.repo.MarkTimeAsSent(time.ID); err != nil {
				slog.Error("error marking time as sent: " + err.Error())
				continue
			}
		}

		time.Sleep(time.Minute)
	}
}

func (s *Service) GetUsers() ([]entities.Employee, error) {
	return s.repo.GetEmployees()
}

func (s *Service) GetProjects(userID string) ([]entities.Project, error) {
	return s.repo.GetProjects(userID)
}

func (s *Service) GetTasks(userID, projectID string) ([]entities.Task, error) {
	return s.repo.GetActiveTasks(userID, projectID)
}

func (s *Service) Actualize() (updated bool, err error) {
	system, err := s.repo.GetSystemInfo()
	if err != nil {
		return false, err
	}

	// fmt.Println("Getting times")
	// times, timesLastUpdate, err := s.external.GetTimes(system.TimesDBLastSynced, "")
	// if err != nil {
	// 	return false, err
	// }
	// fmt.Println("Times last update: ", timesLastUpdate)

	fmt.Println("Getting employees")
	employees, _, err := s.external.GetEmployees(system.EmployeeDBLastSynced)
	if err != nil {
		return false, err
	}
	if err := s.repo.SetEmployees(employees); err != nil {
		return false, err
	}

	fmt.Println("Getting projects")
	projects, projectsLastUpdate, err := s.external.GetProjects(system.ProjectsDBLastSynced)
	if err != nil {
		return false, err
	}
	if err := s.repo.SetProjects(projects); err != nil {
		return false, err
	}

	fmt.Println("Getting not correct person times")
	times, _, err := s.external.GetNotCorrectPersonTimes()
	if err != nil {
		return false, err
	}
	if err := s.SetProfileInTimes(times); err != nil {
		return false, err
	}

	fmt.Println("Getting tasks")
	tasks, tasksLastUpdate, err := s.external.GetTasks("last_edited_time", system.TasksDBLastSynced, "", false)
	if err != nil {
		return false, err
	}

	if err := s.repo.SetTasks(tasks); err != nil {
		return false, err
	}

	system.EmployeeDBLastSynced = 0
	system.ProjectsDBLastSynced = projectsLastUpdate
	if tasksLastUpdate > 0 {
		system.TasksDBLastSynced = tasksLastUpdate
	}

	if err := s.repo.SetSystemInfo(system); err != nil {
		return false, err
	}

	return len(employees) > 0 || len(projects) > 0 || len(tasks) > 0, nil
}

func (s *Service) SetProfileInTimes(times []entities.Time) error {
	for _, time := range times {
		employee, err := s.repo.GetEmployeeByID(time.EmployeeID)
		if err != nil {
			return err
		}
		if err := s.external.SetProfileInTime(time.ID, employee.ProfileID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) WriteOfTime(time *entities.TimeMsg) error {
	return s.repo.SaveTimeWriteOf(time)
}

var forbiddenWords = []string{
	"Фикс",
	"Пофиксить",
	"Фиксить",
	"Правка",
	"Править",
	"Поправить",
	"Исправить",
	"Правки",
	"Исправление",
	"Баг",
	"Безуспешно",
	"Разобраться",
}

func containsForbiddenWord(input string) (string, bool) {
	lowerInput := strings.ToLower(input)
	for _, word := range forbiddenWords {
		if strings.Contains(lowerInput, strings.ToLower(word)) {
			return word, true
		}
	}
	return "", false
}

// TODO: replace with outbox pattern
func (s *Service) ValidateTimes(times []entities.Time) []entities.Row {
	invalidTimes := []entities.Row{}
	for _, time := range times {
		if word, contains := containsForbiddenWord(time.Description); contains {
			time.Description = strings.ReplaceAll(time.Description, strings.ToLower(word), "<b><i>"+strings.ToLower(word)+"</i></b>")
			time.Description = strings.ReplaceAll(time.Description, word, "<b><i>"+word+"</i></b>")
			invalidTimes = append(invalidTimes, time.ToRow())
		}
	}
	return invalidTimes
}

func (s *Service) ValidateTasks(tasks []entities.Task) []entities.Row {
	invalidTasks := []entities.Row{}
	for _, task := range tasks {
		if word, contains := containsForbiddenWord(task.Title); contains {
			task.Title = strings.ReplaceAll(task.Title, strings.ToLower(word), "<b><i>"+strings.ToLower(word)+"</i></b>")
			task.Title = strings.ReplaceAll(task.Title, word, "<b><i>"+word+"</i></b>")
			invalidTasks = append(invalidTasks, task.ToRow())
		}
	}

	return invalidTasks
}

func (s *Service) NotifyEmployeesAboutSalary(ctx context.Context) error {
	employees, err := s.repo.GetEmployeesByNotificationFlag(ctx, entities.NotificationFlagFinance)
	if err != nil {
		slog.Error("error getting employees by notification flag: " + err.Error())
		return err
	}
	for _, employee := range employees {
		if employee.TelegramID != 0 {
			if err := s.external.SendSalaryNotification(ctx, employee.TelegramID); err != nil {
				continue
			}
		}
	}
	return nil

}

func (s *Service) CreateMindmapTasks(mindmapData string) error {
	projectName, tasks, err := mindmap.ParseMarkdownTasks(mindmapData)
	if err != nil {
		slog.Error("Failed to parse tasks from mindmap", slog.String("error", err.Error()))
		return err
	}

	if err := s.external.CreateMindmapTasks(projectName, tasks); err != nil {
		slog.Error("Failed to create tasks in Notion", slog.String("error", err.Error()))
		return err
	}

	return nil
}

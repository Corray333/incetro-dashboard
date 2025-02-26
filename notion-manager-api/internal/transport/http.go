// @title Task Tracker API
// @version 1.0
// @description API for task tracking using notion
// @BasePath /tracker

package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/Corray333/employee_dashboard/internal/entities"
	"github.com/Corray333/employee_dashboard/pkg/auth"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/spf13/viper"
)

type Transport struct {
	router  *chi.Mux
	service service
}

type service interface {
	GetUsers() ([]entities.Employee, error)
	GetProjects(userID string) ([]entities.Project, error)
	GetTasks(userID, projectID string) ([]entities.Task, error)
	WriteOfTime(time *entities.TimeMsg) error

	GetTasksOfEmployee(employee_id string, period_start, period_end int64) ([]entities.Task, error)

	UpdateGoogleSheets() error

	CreateMindmapTasks(mindmap string) error
	GetQuarterTasks() ([]entities.Task, error)

	NotifyEmployeesAboutSalary(ctx context.Context) error

	GetUserRole(username string) entities.DashboardRole
}

func New(service service) *Transport {
	router := NewRouter()

	return &Transport{
		service: service,
		router:  router,
	}
}

func NewRouter() *chi.Mux {
	router := chi.NewMux()
	router.Use(middleware.Logger)

	router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   viper.GetStringSlice("server.cors.allowed_origins"),
		AllowedMethods:   viper.GetStringSlice("server.cors.allowed_methods"),
		AllowedHeaders:   viper.GetStringSlice("server.cors.allowed_headers"),
		AllowCredentials: viper.GetBool("server.cors.allow_credentials"),
		MaxAge:           viper.GetInt("server.cors.max_age"),
	}))

	return router
}

func (s *Transport) Run() {
	slog.Info("Server is starting...")
	panic(http.ListenAndServe("0.0.0.0:"+viper.GetString("server.port"), s.router))
}

func (t *Transport) RegisterRoutes() {

	t.router.Group(func(r chi.Router) {
		r.Use(auth.NewTelegramCredentialsMiddleware())
		r.Use(t.NewDashboardAuthMiddleware())
		r.Use(t.NewDashboardAdminAuthMiddleware())

		r.Get("/api/access", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		r.Get("/api/tasks/employee/{employee_username}", t.getTasksOfEmployee)
		r.Post("/api/update-sheets", t.updateGoogleSheets)
		r.Post("/api/mindmap", t.parseMindmap)
		r.Get("/api/quarter-tasks", t.getQuarterTasks)
		r.Post("/api/salary-notify", t.notifyEmployeesAboutSalary)
	})

	t.router.Group(func(r chi.Router) {
		r.Use(NewTaskTrackerAuthMiddleware())
		r.Get("/api/tracker/employees", t.getEmployees)
		r.Get("/api/tracker/projects", t.getProjects)
		r.Get("/api/tracker/tasks", t.getTasks)
		r.Post("/api/tracker/time", t.writeOfTime)
	})

	// s.router.Get("/tracker/swagger/*", httpSwagger.WrapHandler)

}

// GetEmployees godoc
// @Summary Get all employees
// @Description Retrieves a list of employees.
// @Tags employees
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {array} entities.Employee
// @Failure 500 {string} string "Internal Server Error"
// @Router /tracker/employees [get]
func (t *Transport) getEmployees(w http.ResponseWriter, r *http.Request) {
	users, err := t.service.GetUsers()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting users: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(users); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding users: %s", err.Error()), http.StatusInternalServerError)
		return
	}
}

// GetProjects godoc
// @Summary Get projects for a specific user
// @Description Retrieves a list of projects for a user by user_id.
// @Tags projects
// @Param user_id query string true "User ID"
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {array} entities.Project
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tracker/projects [get]
func (t *Transport) getProjects(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	projects, err := t.service.GetProjects(userID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting projects: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(projects); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding projects: %s", err.Error()), http.StatusInternalServerError)
		return
	}

}

// GetTasks godoc
// @Summary Get tasks for a specific project and user
// @Description Retrieves a list of tasks for a user and project by user_id and project_id.
// @Tags tasks
// @Param user_id query string true "User ID"
// @Param project_id query string true "Project ID"
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {array} entities.Task "List of tasks"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tracker/tasks [get]
func (t *Transport) getTasks(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	projectID := r.URL.Query().Get("project_id")
	if userID == "" {
		http.Error(w, "user_id is required", http.StatusBadRequest)
		return
	}

	if projectID == "" {
		http.Error(w, "project_id is required", http.StatusBadRequest)
		return
	}

	tasks, err := t.service.GetTasks(userID, projectID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting tasks: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding tasks: %s", err.Error()), http.StatusInternalServerError)
		return
	}
}

// WriteOfTime godoc
// @Summary Record the time spent on a task
// @Description Writes the time spent on a task.
// @Tags time
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Param time body entities.TimeMsg true "Time data"
// @Success 201 {string} string "Created"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tracker/time [post]
func (t *Transport) writeOfTime(w http.ResponseWriter, r *http.Request) {
	var time entities.TimeMsg
	if err := json.NewDecoder(r.Body).Decode(&time); err != nil {
		http.Error(w, fmt.Sprintf("Error decoding time: %s", err.Error()), http.StatusBadRequest)
		return
	}

	if err := t.service.WriteOfTime(&time); err != nil {
		http.Error(w, fmt.Sprintf("Error writing time: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func NewTaskTrackerAuthMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			authToken := r.Header.Get("Authorization")
			authToken = strings.TrimPrefix(authToken, "Bearer ")
			if authToken != os.Getenv("AUTH_TOKEN") {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

type UserCredentials interface {
	GetUserID() int64
	GetUsername() string
}

func (t *Transport) NewDashboardAuthMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			credsI := r.Context().Value(entities.ContextKeyUserCredentials)
			if credsI == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			creds, ok := credsI.(UserCredentials)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			role := t.service.GetUserRole(creds.GetUsername())

			r = r.WithContext(context.WithValue(r.Context(), entities.ContextKeyUserRole, role))
			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

func (t *Transport) NewDashboardAdminAuthMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			roleI := r.Context().Value(entities.ContextKeyUserRole)
			if roleI == nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			role, ok := roleI.(entities.DashboardRole)
			if !ok {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			if role != entities.DashboardRoleAdmin {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

// GetTasksOfEmployee godoc
// @Summary Get tasks of a specific employee within a period
// @Description Retrieves a list of tasks for an employee by employee_id and a time period.
// @Tags tasks
// @Param employee_id path string true "Employee ID"
// @Param period_start query int64 true "Period Start"
// @Param period_end query int64 true "Period End"
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {array} entities.Task "List of tasks"
// @Failure 400 {string} string "Bad Request"
// @Failure 500 {string} string "Internal Server Error"
// @Router /tasks/employee/{employee_id} [get]
func (t *Transport) getTasksOfEmployee(w http.ResponseWriter, r *http.Request) {
	employeeUsername := chi.URLParam(r, "employee_username")
	periodStartStr := r.URL.Query().Get("period_start")
	periodEndStr := r.URL.Query().Get("period_end")

	if employeeUsername == "" {
		http.Error(w, "employee_username is required", http.StatusBadRequest)
		return
	}

	periodStart, err := strconv.ParseInt(periodStartStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid period_start", http.StatusBadRequest)
		return
	}

	periodEnd, err := strconv.ParseInt(periodEndStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid period_end", http.StatusBadRequest)
		return
	}

	tasks, err := t.service.GetTasksOfEmployee(employeeUsername, periodStart, periodEnd)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting tasks: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding tasks: %s", err.Error()), http.StatusInternalServerError)
		return
	}
}

// UpdateGoogleSheets godoc
// @Summary Update Google Sheets with the latest data
// @Description Updates Google Sheets with the latest data from the system.
// @Tags sheets
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {string} string "Sheets updated successfully"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/update-sheets [post]
func (t *Transport) updateGoogleSheets(w http.ResponseWriter, r *http.Request) {
	if err := t.service.UpdateGoogleSheets(); err != nil {
		http.Error(w, fmt.Sprintf("Error updating Google Sheets: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Sheets updated successfully"))
}

func (t *Transport) parseMindmap(w http.ResponseWriter, r *http.Request) {
	file, _, err := r.FormFile("file")
	if err != nil {
		slog.Error(err.Error())
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		slog.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err := t.service.CreateMindmapTasks(string(data)); err != nil {
		slog.Error(err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetQuarterTasks godoc
// @Summary Get tasks for the current quarter
// @Description Retrieves a list of tasks for the current quarter.
// @Tags tasks
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {array} entities.Task "List of tasks"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/quarter-tasks [get]
func (t *Transport) getQuarterTasks(w http.ResponseWriter, r *http.Request) {
	tasks, err := t.service.GetQuarterTasks()
	if err != nil {
		http.Error(w, fmt.Sprintf("Error getting quarter tasks: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(tasks); err != nil {
		http.Error(w, fmt.Sprintf("Error encoding tasks: %s", err.Error()), http.StatusInternalServerError)
		return
	}
}

// NotifyEmployeesAboutSalary godoc
// @Summary Notify employees about their salary
// @Description Sends notifications to employees about their salary.
// @Tags notifications
// @Produce json
// @Param Authorization header string true "Bearer token"
// @Success 200 {string} string "Notifications sent successfully"
// @Failure 500 {string} string "Internal Server Error"
// @Router /api/salary-notify [get]
func (t *Transport) notifyEmployeesAboutSalary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := t.service.NotifyEmployeesAboutSalary(ctx); err != nil {
		http.Error(w, fmt.Sprintf("Error notifying employees about salary: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Notifications sent successfully"))
}

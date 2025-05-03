package transport

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	entity_task "github.com/Corray333/employee_dashboard/internal/domains/task/entities/task"
	"github.com/Corray333/employee_dashboard/internal/transport"
	"github.com/go-chi/chi/v5"
)

type service interface {
	CreateTask(ctx context.Context, task *entity_task.TaskOutboxMsg) error
}
type TaskTransport struct {
	service service
	router  *chi.Mux
}

func NewTaskTransport(router *chi.Mux, service service) *TaskTransport {
	t := &TaskTransport{
		service: service,
		router:  router,
	}

	return t
}

func (t *TaskTransport) RegisterRoutes() {
	t.router.Group(func(r chi.Router) {
		r.Use(transport.NewTaskTrackerAuthMiddleware())

		r.Post("/api/task", t.createTask)
	})
}

func (t *TaskTransport) createTask(w http.ResponseWriter, r *http.Request) {
	req := &entity_task.TaskOutboxMsg{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		slog.Error("Error decoding request", "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := t.service.CreateTask(r.Context(), req); err != nil {
		slog.Error("Error creating task", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

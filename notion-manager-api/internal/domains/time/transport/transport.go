package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	entity_time "github.com/Corray333/employee_dashboard/internal/domains/time/entities/time"
	"github.com/Corray333/employee_dashboard/internal/transport"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type service interface {
	CreateTimeWriteOf(ctx context.Context, time *entity_time.TimeOutboxMsg) error
}
type TimeTransport struct {
	service service
	router  *chi.Mux
}

func NewTimeTransport(router *chi.Mux, service service) *TimeTransport {
	t := &TimeTransport{
		service: service,
		router:  router,
	}

	return t
}

func (t *TimeTransport) RegisterRoutes() {
	t.router.Group(func(r chi.Router) {
		r.Use(transport.NewTaskTrackerAuthMiddleware())
		fmt.Println("Register time routes")
		r.Post("/api/time", t.writeOfTime)
	})
}

func (t *TimeTransport) writeOfTime(w http.ResponseWriter, r *http.Request) {
	req := &entity_time.TimeOutboxMsg{}
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if req.TaskID == uuid.Nil || req.EmployeeID == uuid.Nil {
		http.Error(w, "task_id and employee_id are required", http.StatusBadRequest)
		return
	}

	if err := t.service.CreateTimeWriteOf(r.Context(), req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)

}

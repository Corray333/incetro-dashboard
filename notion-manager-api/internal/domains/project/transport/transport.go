package transport

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Corray333/employee_dashboard/internal/domains/project/entities/project"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type service interface {
	UpdateProjectSheets(ctx context.Context, projectID uuid.UUID) error

	ListProjectsWithLinkedSheets(ctx context.Context) ([]project.Project, error)
}
type ProjectTransport struct {
	service service
	router  *chi.Mux
}

func NewProjectTransport(router *chi.Mux, service service) *ProjectTransport {
	t := &ProjectTransport{
		service: service,
		router:  router,
	}

	return t
}

func (t *ProjectTransport) RegisterRoutes() {
	t.router.Group(func(r chi.Router) {
		// r.Use(transport.NewTaskTrackerAuthMiddleware())

		r.Post("/api/projects/{projectID}/update-sheets", t.updateSheets)
		r.Get("/api/projects/with-sheets", t.listProjectsWithLinkedSheets)
	})
}

func (t *ProjectTransport) updateSheets(w http.ResponseWriter, r *http.Request) {
	projectIDStr := chi.URLParam(r, "projectID")
	projectID, err := uuid.Parse(projectIDStr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := t.service.UpdateProjectSheets(r.Context(), projectID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (t *ProjectTransport) listProjectsWithLinkedSheets(w http.ResponseWriter, r *http.Request) {
	projects, err := t.service.ListProjectsWithLinkedSheets(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := json.NewEncoder(w).Encode(projects); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)

}

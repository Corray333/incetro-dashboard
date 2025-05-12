package postgres

import (
	"context"
	"log/slog"

	entity_time "github.com/Corray333/employee_dashboard/internal/domains/time/entities/time"
	"github.com/google/uuid"
)

func (r *TimePostgresRepository) ListTimesByProject(ctx context.Context, projectID uuid.UUID, offset int, limit int) ([]entity_time.Time, error) {
	times := []timeDB{}
	if err := r.DB().Select(&times, `
		SELECT 
			times.*, 
			COALESCE(projects.name, '') as project_name, 
			COALESCE(tasks.title, '') as task_name, 
			COALESCE(expertise.name, '') as expertise,
			COALESCE(tasks.estimate, 0) as task_estimate,
			COALESCE(employees.username, '') as who_did
		FROM times
		LEFT JOIN projects ON times.project_id = projects.project_id::uuid
		LEFT JOIN tasks ON times.task_id = tasks.task_id::uuid
		LEFT JOIN expertise ON times.expertise_id = expertise.expertise_id::uuid
		LEFT JOIN employees ON employees.employee_id::uuid = times.employee_id
		WHERE times.project_id = $3
		OFFSET $1 LIMIT $2
	`, offset, limit, projectID); err != nil {
		slog.Error("Error while getting times", "error", err)
		return nil, err
	}
	result := []entity_time.Time{}
	for _, time := range times {
		result = append(result, *time.ToEntity())
	}

	return result, nil
}

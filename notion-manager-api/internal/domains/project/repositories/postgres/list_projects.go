package postgres

import (
	"context"
	"log/slog"

	"github.com/Corray333/employee_dashboard/internal/domains/project/entities/project"
	"github.com/google/uuid"
)

type projectPostgres struct {
	ID         uuid.UUID `db:"project_id"`
	Name       string    `db:"name"`
	Icon       string    `db:"icon"`
	IconType   string    `db:"icon_type"`
	Status     string    `db:"status"`
	Type       string    `db:"type"`
	ManagerID  uuid.UUID `db:"manager_id"`
	SheetsLink string    `db:"sheets_link"`
}

func (p *projectPostgres) ToEntity() project.Project {
	return project.Project{
		ID:         p.ID,
		Name:       p.Name,
		Icon:       p.Icon,
		IconType:   p.IconType,
		Status:     p.Status,
		Type:       p.Type,
		ManagerID:  p.ManagerID,
		SheetsLink: p.SheetsLink,
	}
}

func (r *ProjectPostgresRepository) ListProjects(ctx context.Context) ([]project.Project, error) {
	var projects []projectPostgres
	if err := r.DB().Select(&projects, "SELECT * FROM projects"); err != nil {
		slog.Error("Error listing projects", "error", err)
		return nil, err
	}

	var result []project.Project
	for _, p := range projects {
		result = append(result, p.ToEntity())
	}
	return result, nil
}

func (r *ProjectPostgresRepository) ListProjectsWithLinkedSheets(ctx context.Context) ([]project.Project, error) {
	var projects []projectPostgres
	if err := r.DB().Select(&projects, "SELECT * FROM projects WHERE sheets_link != ''"); err != nil {
		slog.Error("Error listing projects with linked sheets", "error", err)
		return nil, err
	}

	var result []project.Project
	for _, p := range projects {
		result = append(result, p.ToEntity())
	}
	return result, nil
}

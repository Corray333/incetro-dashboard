package postgres

import (
	"context"
	"log/slog"

	"github.com/Corray333/employee_dashboard/internal/domains/project/entities/project"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type projectPostgres struct {
	ID         uuid.UUID  `db:"project_id"`
	Name       string     `db:"name"`
	Icon       string     `db:"icon"`
	IconType   string     `db:"icon_type"`
	Status     string     `db:"status"`
	Type       string     `db:"type"`
	ManagerID  uuid.UUID  `db:"manager_id"`
	SheetsLink string     `db:"sheets_link"`
	ClientID   *uuid.UUID `db:"client_id"`
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
		ClientID:   p.ClientID,
	}
}

func (r *ProjectPostgresRepository) ListProjects(ctx context.Context) ([]project.Project, error) {
	var projects []projectPostgres
	query := `
		SELECT 
			p.project_id,
			p.name,
			p.icon,
			p.icon_type,
			p.status,
			p.type,
			p.manager_id,
			p.sheets_link,
			c.client_id
		FROM projects p
		LEFT JOIN clients c ON p.project_id::text = ANY(c.project_ids)
		ORDER BY p.project_id
	`
	if err := r.DB().Select(&projects, query); err != nil {
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
	query := `
		SELECT 
			p.project_id,
			p.name,
			p.icon,
			p.icon_type,
			p.status,
			p.type,
			p.manager_id,
			p.sheets_link,
			c.client_id
		FROM projects p
		LEFT JOIN clients c ON p.project_id::text = ANY(c.project_ids)
		WHERE p.sheets_link != ''
		ORDER BY p.project_id
	`
	if err := r.DB().Select(&projects, query); err != nil {
		slog.Error("Error listing projects with linked sheets", "error", err)
		return nil, err
	}

	var result []project.Project
	for _, p := range projects {
		result = append(result, p.ToEntity())
	}
	return result, nil
}

func (r *ProjectPostgresRepository) GetProjectsByIDs(ctx context.Context, projectIDs []uuid.UUID) ([]project.Project, error) {
	if len(projectIDs) == 0 {
		return []project.Project{}, nil
	}

	var projects []projectPostgres
	query, args, err := sqlx.In("SELECT * FROM projects WHERE project_id IN (?)", projectIDs)
	if err != nil {
		slog.Error("Error building query for projects by IDs", "error", err)
		return nil, err
	}

	// Rebind the query to use PostgreSQL-style placeholders ($1, $2, ...)
	query = r.DB().Rebind(query)

	if err := r.DB().Select(&projects, query, args...); err != nil {
		slog.Error("Error getting projects by IDs", "error", err)
		return nil, err
	}

	var result []project.Project
	for _, p := range projects {
		result = append(result, p.ToEntity())
	}
	return result, nil
}

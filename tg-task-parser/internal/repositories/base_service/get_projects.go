package base_service

import (
	"context"
	"log/slog"

	"github.com/corray333/tg-task-parser/internal/entities/project"
)

func (r *BaseService) GetProjects(ctx context.Context) ([]project.Project, error) {
	projects := []project.Project{}
	if err := r.db.Select(&projects, "SELECT project_id, name FROM projects WHERE status = $1", project.StatusActive); err != nil {
		slog.Error("Error while getting projects", "error", err)
		return nil, err
	}
	return projects, nil
}

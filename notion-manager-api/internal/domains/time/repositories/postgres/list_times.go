package postgres

import (
	"context"
	"log/slog"

	entity_time "github.com/Corray333/employee_dashboard/internal/domains/time/entities/time"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *TimePostgresRepository) ListTimes(ctx context.Context, filter entity_time.TimeFilter, offset int, limit int) ([]entity_time.Time, error) {

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).
		Select(
			"times.*",
			"COALESCE(projects.name, '') AS project_name",
			"COALESCE(tasks.title, '') AS task_name",
			"COALESCE(expertise.name, '') AS expertise",
			"COALESCE(tasks.estimate, 0) AS task_estimate",
			"COALESCE(employees.username, '') AS who_did",
		).
		From("times").
		LeftJoin("projects ON times.project_id = projects.project_id::uuid").
		LeftJoin("tasks ON times.task_id = tasks.task_id::uuid").
		LeftJoin("expertise ON times.expertise_id = expertise.expertise_id::uuid").
		LeftJoin("employees ON employees.employee_id::uuid = times.employee_id").
		Offset(uint64(offset)).
		Limit(uint64(limit))

	// Применение фильтра по ProjectID, если он задан
	if filter.ProjectID != uuid.Nil {
		builder = builder.Where(squirrel.Eq{"times.project_id": filter.ProjectID})
	}

	// Преобразование билдера в SQL-запрос
	query, args, err := builder.ToSql()
	if err != nil {
		slog.Error("Ошибка при построении SQL-запроса", "error", err)
		return nil, err
	}

	// Выполнение запроса
	var times []timeDB
	if err := r.DB().Select(&times, query, args...); err != nil {
		slog.Error("Ошибка при выполнении запроса к базе данных", "error", err)
		return nil, err
	}

	// Преобразование результатов в сущности
	result := make([]entity_time.Time, 0, len(times))
	for _, t := range times {
		result = append(result, *t.ToEntity())
	}

	return result, nil
}

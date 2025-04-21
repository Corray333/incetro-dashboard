package postgres

import (
	"context"
	"log/slog"

	"github.com/Corray333/employee_dashboard/internal/domains/feedback/entities/feedback"
	"github.com/Masterminds/squirrel"
	"github.com/google/uuid"
)

func (r *FeedbackPostgresRepository) ListFeedbacks(ctx context.Context, filter *feedback.Feedback) ([]feedback.Feedback, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)

	query := psql.
		Select("*").
		From("feedbacks")

	if filter.ProjectID != uuid.Nil {
		query = query.Where(squirrel.Eq{"project": filter.ProjectID})
	}

	if filter.Status != "" {
		query = query.Where(squirrel.Eq{"status": filter.Status})
	}

	sqlQuery, args, err := query.ToSql()
	if err != nil {
		slog.Error("Error building SQL query", "error", err)
		return nil, err
	}

	var feedbacks []feedbackPostgres
	err = r.DB().SelectContext(ctx, &feedbacks, sqlQuery, args...)
	if err != nil {
		slog.Error("Error executing SQL query", "error", err)
		return nil, err
	}

	var result []feedback.Feedback
	for _, f := range feedbacks {
		result = append(result, *f.ToEntity())
	}

	return result, nil
}

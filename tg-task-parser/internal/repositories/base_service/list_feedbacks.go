package base_service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/corray333/tg-task-parser/internal/entities/feedback"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type feedbackPostgres struct {
	ID          uuid.UUID `db:"feedback_id"`
	Text        string    `db:"text"`
	Type        string    `db:"type"`
	Priority    string    `db:"priority"`
	TaskID      uuid.UUID `db:"task_id"`
	ProjectID   uuid.UUID `db:"project_id"`
	CreatedDate time.Time `db:"created_date"`
	Direction   string    `db:"direction"`
	Status      string    `db:"status"`
}

func (f *feedbackPostgres) ToEntity() *feedback.Feedback {
	return &feedback.Feedback{
		ID:          f.ID,
		Text:        f.Text,
		Type:        f.Type,
		Priority:    f.Priority,
		TaskID:      f.TaskID,
		ProjectID:   f.ProjectID,
		CreatedDate: f.CreatedDate,
		Direction:   f.Direction,
		Status:      feedback.Status(f.Status),
	}
}

func (r *BaseService) ListFeedbacks(ctx context.Context, projectID uuid.UUID, statuses []feedback.Status) ([]feedback.Feedback, error) {
	var feedbacks []feedbackPostgres

	query, args, err := sqlx.In("SELECT * FROM feedbacks WHERE project_id = ? AND status IN (?)", projectID, statuses)
	if err != nil {
		slog.Error("Error building SQL query", "error", err)
		return nil, err
	}

	// Rebind the query to use PostgreSQL-style placeholders ($1, $2, ...)
	query = r.db.Rebind(query)

	fmt.Println(query, args)
	if err := r.db.Select(&feedbacks, query, args...); err != nil {
		slog.Error("Error executing SQL query", "error", err)
		return nil, err
	}

	var result []feedback.Feedback
	for _, f := range feedbacks {
		result = append(result, *f.ToEntity())
	}

	return result, nil
}

package postgres

import (
	"context"
	"log/slog"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/feedback/entities/feedback"
	"github.com/google/uuid"
)

// ID          uuid.UUID `json:"id"`
// Text        string    `json:"text"`
// Type        string    `json:"type"`
// Priority    string    `json:"priority"`
// TaskID      uuid.UUID `json:"task"`
// ProjectID   uuid.UUID `json:"project"`
// CreatedDate time.Time `json:"createdDate"`
// Direction   string    `json:"direction"`
// Status      string    `json:"status"`

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
		Status:      f.Status,
	}
}

func (r *FeedbackPostgresRepository) SetFeedback(ctx context.Context, feedback *feedback.Feedback) error {
	feedbackPostgres := &feedbackPostgres{
		ID:          feedback.ID,
		Text:        feedback.Text,
		Type:        feedback.Type,
		Priority:    feedback.Priority,
		TaskID:      feedback.TaskID,
		ProjectID:   feedback.ProjectID,
		CreatedDate: feedback.CreatedDate,
		Direction:   feedback.Direction,
		Status:      feedback.Status,
	}
	_, err := r.DB().NamedExec(`
		INSERT INTO feedbacks (feedback_id, text, type, priority, task_id, project_id, created_date, direction, status)
		VALUES (:feedback_id, :text, :type, :priority, :task_id, :project_id, :created_date, :direction, :status)
		ON CONFLICT (feedback_id) DO UPDATE SET
			text = :text,
			type = :type,
			priority = :priority,
			task_id = :task_id,
			project_id = :project_id,
			created_date = :created_date,
			direction = :direction,
			status = :status
	`, feedbackPostgres)
	if err != nil {
		slog.Error("Error setting feedback", "error", err)
		return err
	}
	return nil
}

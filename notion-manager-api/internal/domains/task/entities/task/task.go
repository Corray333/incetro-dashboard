package task

import (
	"time"

	"github.com/google/uuid"
)

type Task struct {
	ID             uuid.UUID `json:"id"`
	CreatedTime    string    `json:"created_time"`
	LastEditedTime string    `json:"last_edited_time"`

	Task        string      `json:"task"`
	Priority    string      `json:"priority"`
	Status      Status      `json:"status"`
	ParentID    uuid.UUID   `json:"parent_id"`
	Responsible []uuid.UUID `json:"responsible_ids"`
	CreatorID   uuid.UUID   `json:"creator_id"`
	ExecutorIDs []uuid.UUID `json:"executor_ids"`
	ProjectID   uuid.UUID   `json:"project_id"`
	Estimate    float64     `json:"estimate"`
	Tags        []string    `json:"tags"`
	Start       time.Time   `json:"start"`
	End         time.Time   `json:"end"`
}

// task, estimate, start, end, executor
type TaskOutboxMsg struct {
	ID         int64     `json:"id"`
	ProjectID  uuid.UUID `json:"projectID"`
	ExecutorID uuid.UUID `json:"executorID"`
	Task       string    `json:"task"`
	Estimate   float64   `json:"estimate"`
	Priority   string    `json:"priority"`
	Start      time.Time `json:"start"`
	End        time.Time `json:"end"`
}

func (t *TaskOutboxMsg) ToEntity() *Task {
	return &Task{Task: t.Task,
		Estimate:    t.Estimate,
		Priority:    t.Priority,
		Start:       t.Start,
		End:         t.End,
		ExecutorIDs: []uuid.UUID{t.ExecutorID},
		ProjectID:   t.ProjectID,
	}
}

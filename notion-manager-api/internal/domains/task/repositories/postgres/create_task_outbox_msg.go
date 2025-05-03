package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	entity_task "github.com/Corray333/employee_dashboard/internal/domains/task/entities/task"
	"github.com/google/uuid"
)

type taskOutboxMsgDB struct {
	ID         int64     `db:"task_msg_id"`
	Task       string    `db:"task"`
	Estimate   float64   `db:"estimate"`
	Priority   string    `db:"priority"`
	Start      time.Time `db:"deadline_start"`
	End        time.Time `db:"deadline_end"`
	ExecutorID uuid.UUID `db:"executor_id"`
	ProjectID  uuid.UUID `db:"project_id"`
}

func (m *taskOutboxMsgDB) toEntity() *entity_task.TaskOutboxMsg {
	return &entity_task.TaskOutboxMsg{
		ID:         m.ID,
		Task:       m.Task,
		Estimate:   m.Estimate,
		Priority:   m.Priority,
		Start:      m.Start,
		End:        m.End,
		ExecutorID: m.ExecutorID,
		ProjectID:  m.ProjectID,
	}
}

func taskOutboxMsgDBFromEntity(msg *entity_task.TaskOutboxMsg) *taskOutboxMsgDB {
	return &taskOutboxMsgDB{
		ID:         msg.ID,
		Task:       msg.Task,
		Estimate:   msg.Estimate,
		Priority:   msg.Priority,
		Start:      msg.Start,
		End:        msg.End,
		ExecutorID: msg.ExecutorID,
		ProjectID:  msg.ProjectID,
	}
}

func (r *TaskPostgresRepository) CreateTaskOutboxMsg(ctx context.Context, msg *entity_task.TaskOutboxMsg) error {
	tx, isNew, err := r.GetTx(ctx)
	if err != nil {
		return err
	}
	if isNew {
		defer tx.Rollback()
	}
	if _, err := tx.NamedExec(`INSERT INTO task_outbox (task, estimate, priority, deadline_start, deadline_end, executor_id, project_id) VALUES (:task, :estimate, :priority, :deadline_start, :deadline_end, :executor_id, :project_id)`, taskOutboxMsgDBFromEntity(msg)); err != nil {
		slog.Error("Error insert task outbox msg", "error", err)
		return err
	}

	if isNew {
		if err := tx.Commit(); err != nil {
			slog.Error("Error commit transaction", "error", err)
			return err
		}
	}

	return nil
}

func (r *TaskPostgresRepository) GetTaskOutboxMsgs(ctx context.Context) ([]entity_task.TaskOutboxMsg, error) {
	msgs := make([]taskOutboxMsgDB, 0)
	if err := r.DB().SelectContext(ctx, &msgs, `SELECT * FROM task_outbox LIMIT 50`); err != nil {
		slog.Error("Error get task outbox msgs", "error", err)
		return nil, err
	}

	result := make([]entity_task.TaskOutboxMsg, 0, len(msgs))
	for _, msg := range msgs {
		result = append(result, *msg.toEntity())
	}

	return result, nil

}

func (r *TaskPostgresRepository) DeleteTaskOutboxMsg(ctx context.Context, msg *entity_task.TaskOutboxMsg) error {
	tx, isNew, err := r.GetTx(ctx)
	if err != nil {
		return err
	}
	if isNew {
		defer tx.Rollback()
	}
	msgDB := taskOutboxMsgDBFromEntity(msg)
	fmt.Printf("Deleting: %+v\n", msg)
	if _, err := tx.NamedExec(`DELETE FROM task_outbox WHERE task_msg_id = :task_msg_id`, msgDB); err != nil {
		slog.Error("Error delete task outbox msg", "error", err)
		return err
	}

	if isNew {
		if err := tx.Commit(); err != nil {
			slog.Error("Error commit transaction", "error", err)
			return err
		}
	}

	return nil
}

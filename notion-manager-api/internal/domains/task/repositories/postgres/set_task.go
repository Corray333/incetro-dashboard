package postgres

import (
	"context"
	"log/slog"
	"time"

	"github.com/Corray333/employee_dashboard/internal/domains/task/entities/task"
	entity_task "github.com/Corray333/employee_dashboard/internal/domains/task/entities/task"
	"github.com/google/uuid"
)

type taskDB struct {
	ID             uuid.UUID `db:"task_id"`
	CreatedTime    time.Time `db:"created_time"`
	LastEditedTime time.Time `db:"last_edited_time"`
	Title          string    `db:"title"`
	ExecutorID     uuid.UUID `db:"executor_id"`
	ResponsibleID  uuid.UUID `db:"responsible_id"`
	Priority       string    `db:"priority"`
	Status         string    `db:"status"`
	ParentID       uuid.UUID `db:"parent_id"`
	CreatorID      uuid.UUID `db:"creator_id"`
	ProjectID      uuid.UUID `db:"project_id"`
	Estimate       float64   `db:"estimate"`
	Start          time.Time `db:"start"`
	End            time.Time `db:"end"`
	PreviousID     uuid.UUID `db:"previous_id"`
	NextID         uuid.UUID `db:"next_id"`
	TotalHours     float64   `db:"total_hours"`
	TBH            float64   `db:"tbh"`
	CP             float64   `db:"cp"`
	TotalEstimate  float64   `db:"total_estimate"`
	PlanFact       float64   `db:"plan_fact"`
	Duration       float64   `db:"duration"`
	CR             float64   `db:"cr"`
	IKP            string    `db:"ikp"`
	MainTask       string    `db:"main_task"`
	Tags           []string  `db:"-"`

	Expertise   string `db:"expertise"` // Get from
	ProjectName string `db:"project_name"`
	ParentName  string `db:"parent_name"`
	Direction   string `db:"direction"`
}

func (r *TaskPostgresRepository) ListTasks(ctx context.Context, limit, offset int) ([]task.Task, error) {
	tasks := make([]taskDB, 0)
	if err := r.DB().SelectContext(ctx, &tasks, `
		SELECT 
			tasks.task_id, tasks.created_time, tasks.last_edited_time, tasks.title, tasks.priority, tasks.status,
			tasks.parent_id AS parent_id, tasks.creator_id, tasks.project_id, tasks.estimate,
			tasks.start, tasks."end", tasks.previous_id, tasks.next_id, tasks.total_hours, tasks.tbh,
			tasks.cp, tasks.total_estimate, tasks.plan_fact, tasks.duration, tasks.cr, tasks.ikp, tasks.main_task,
			tasks.executor_id, tasks.responsible_id,
			COALESCE(t.title, '') AS parent_name,
			COALESCE(exp.name, '') AS expertise
		FROM tasks
		LEFT JOIN tasks t ON t.task_id = tasks.parent_id
		LEFT JOIN employees e ON e.employee_id = tasks.executor_id
		LEFT JOIN expertise exp ON exp.expertise_id = e.expertise_id
		LIMIT $1 OFFSET $2
	`, limit, offset); err != nil {
		slog.Error("Error listing tasks", "error", err)
		return nil, err
	}

	result := make([]task.Task, 0, len(tasks))
	for _, t := range tasks {
		if err := r.DB().Select(&t.Tags, `SELECT tag FROM task_tag WHERE task_id = $1`, t.ID); err != nil {
			slog.Error("Error getting task tags", "error", err)
			return nil, err
		}
		result = append(result, *t.toEntity())
	}

	return result, nil
}

func entityToTaskDB(task *entity_task.Task) *taskDB {
	return &taskDB{
		ID:             task.ID,
		CreatedTime:    task.CreatedTime,
		LastEditedTime: task.LastEditedTime,
		Title:          task.Task,
		Priority:       task.Priority,
		Status:         string(task.Status),
		ParentID:       task.ParentID,
		CreatorID:      task.CreatorID,
		ProjectID:      task.ProjectID,
		Estimate:       task.Estimate,
		Start:          task.Start,
		End:            task.End,
		PreviousID:     task.PreviousID,
		NextID:         task.NextID,
		TotalHours:     task.TotalHours,
		TBH:            task.TBH,
		CP:             task.CP,
		TotalEstimate:  task.TotalEstimate,
		PlanFact:       task.PlanFact,
		Duration:       task.Duration,
		CR:             task.CR,
		IKP:            task.IKP,
		MainTask:       task.MainTask,

		Expertise:     task.Expertise,
		ProjectName:   task.ProjectName,
		ParentName:    task.ParentName,
		Direction:     task.Direction,
		ExecutorID:    task.ExecutorID,
		ResponsibleID: task.ResponsibleID,
	}
}

func (r *TaskPostgresRepository) SetTask(ctx context.Context, task *entity_task.Task) error {
	tx, isNew, err := r.GetTx(ctx)
	if err != nil {
		return err
	}
	if isNew {
		defer tx.Rollback()
	}

	taskDB := entityToTaskDB(task)

	_, err = tx.NamedExec(`
		INSERT INTO tasks (
			task_id, created_time, last_edited_time, title, priority, status, parent_id,
			creator_id, project_id, estimate, start, "end", previous_id, next_id,
			total_hours, tbh, cp, total_estimate, plan_fact, duration, cr, ikp, main_task, executor_id, responsible_id
		) VALUES (
			:task_id, :created_time, :last_edited_time, :title, :priority, :status, :parent_id,
			:creator_id, :project_id, :estimate, :start, :end, :previous_id, :next_id,
			:total_hours, :tbh, :cp, :total_estimate, :plan_fact, :duration, :cr, :ikp, :main_task, :executor_id, :responsible_id
		)
		ON CONFLICT (task_id) DO UPDATE SET
			created_time = :created_time,
			last_edited_time = :last_edited_time,
			title = :title,
			priority = :priority,
			status = :status,
			parent_id = :parent_id,
			creator_id = :creator_id,
			project_id = :project_id,
			estimate = :estimate,
			start = :start,
			"end" = :end,
			previous_id = :previous_id,
			next_id = :next_id,
			total_hours = :total_hours,
			tbh = :tbh,
			cp = :cp,
			total_estimate = :total_estimate,
			plan_fact = :plan_fact,
			duration = :duration,
			cr = :cr,
			ikp = :ikp,
			main_task = :main_task,
			executor_id = :executor_id,
			responsible_id = :responsible_id;
	`, taskDB)
	if err != nil {
		slog.Error("Error setting task", "error", err)
		return err
	}

	for _, tag := range task.Tags {
		_, err = tx.Exec(`INSERT INTO task_tag (task_id, tag) VALUES ($1, $2) ON CONFLICT DO NOTHING`, taskDB.ID, tag)
		if err != nil {
			slog.Error("Error setting task tag", "error", err)
			return err
		}
	}

	if isNew {
		return tx.Commit()
	}

	return nil
}

func (t *taskDB) toEntity() *task.Task {

	tags := make([]task.Tag, 0, len(t.Tags))
	for _, tag := range t.Tags {
		tags = append(tags, task.Tag(tag))
	}

	return &task.Task{
		ID:             t.ID,
		CreatedTime:    t.CreatedTime,
		LastEditedTime: t.LastEditedTime,
		Task:           t.Title,
		ExecutorID:     t.ExecutorID,
		ResponsibleID:  t.ResponsibleID,
		Priority:       t.Priority,
		Status:         task.Status(t.Status),
		ParentID:       t.ParentID,
		CreatorID:      t.CreatorID,
		ProjectID:      t.ProjectID,
		Estimate:       t.Estimate,
		Start:          t.Start,
		End:            t.End,
		PreviousID:     t.PreviousID,
		NextID:         t.NextID,
		TotalHours:     t.TotalHours,
		TBH:            t.TBH,
		CP:             t.CP,
		TotalEstimate:  t.TotalEstimate,
		PlanFact:       t.PlanFact,
		Duration:       t.Duration,
		CR:             t.CR,
		IKP:            t.IKP,
		MainTask:       t.MainTask,
		Tags:           tags,

		Expertise:   t.Expertise,
		ProjectName: t.ProjectName,
		ParentName:  t.ParentName,
		Direction:   t.Direction,
	}
}

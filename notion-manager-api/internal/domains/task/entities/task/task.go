package task

import (
	"slices"
	"time"

	"github.com/google/uuid"
)

type Tag string

const (
	TagIOS        Tag = "iOS"
	TagAndroid    Tag = "Android"
	TagFrontend   Tag = "Frontend"
	TagBackend    Tag = "Backend"
	TagQA         Tag = "QA"
	TagFlutter    Tag = "Flutter"
	TagUIUX       Tag = "UI/UX"
	TagManagement Tag = "Management"
)

var DirectionsTags = []Tag{
	TagIOS,
	TagAndroid,
	TagFrontend,
	TagBackend,
	TagQA,
	TagFlutter,
	TagUIUX,
	TagManagement,
}

type Task struct {
	ID             uuid.UUID `json:"id"`
	CreatedTime    time.Time `json:"created_time"`
	LastEditedTime time.Time `json:"last_edited_time"`

	Task          string    `json:"task"`
	Priority      string    `json:"priority"`
	Status        Status    `json:"status"`
	ParentID      uuid.UUID `json:"parent_id"`
	ResponsibleID uuid.UUID `json:"responsible_id"`
	CreatorID     uuid.UUID `json:"creator_id"`
	ExecutorID    uuid.UUID `json:"executor_id"`
	ProjectID     uuid.UUID `json:"project_id"`
	Estimate      float64   `json:"estimate"`
	Tags          []Tag     `json:"tags"`
	Start         time.Time `json:"start"`
	End           time.Time `json:"end"`
	SH            float64   `json:"sh"`

	PreviousID    uuid.UUID   `json:"previous_ids"`
	NextID        uuid.UUID   `json:"next_ids"`
	TotalHours    float64     `json:"total_hours"`
	Subtasks      []uuid.UUID `json:"subtasks_ids"`
	TBH           float64     `json:"tbh"`
	CP            float64     `json:"cp"`
	TotalEstimate float64     `json:"total_estimate"`
	PlanFact      float64     `json:"plan_fact"`
	Duration      float64     `json:"duration"`
	CR            float64     `json:"cr"`
	IKP           string      `json:"ikp"`
	MainTask      string      `json:"main_task"`

	ParentName  string `json:"parent_name"`
	ProjectName string `json:"project_name"`
	Expertise   string `json:"expertise"`
	Direction   string `json:"direction"`
}

func (t *Task) GetDirection() string {
	for _, tag := range t.Tags {
		if slices.Contains(DirectionsTags, tag) {
			return string(tag)
		}
	}

	return ""
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
	return &Task{
		Task:       t.Task,
		Estimate:   t.Estimate,
		Priority:   t.Priority,
		Start:      t.Start,
		End:        t.End,
		ExecutorID: t.ExecutorID,
		ProjectID:  t.ProjectID,
	}
}

package postgres

import (
	"context"
	"log/slog"
	"time"

	entity_time "github.com/Corray333/employee_dashboard/internal/domains/time/entities/time"

	"github.com/google/uuid"
)

// type entity_time.Time struct {
// 	ID            uuid.UUID `json:"id"`
// 	TotalHours    float64   `json:"total_hours"`
// 	PayableHours  float64   `json:"payable_hours"`
// 	TaskID        uuid.UUID `json:"task_id"`
// 	Direction     string    `json:"direction"`
// 	WorkDate      time.Time `json:"work_date"`
// 	EmployeeID    uuid.UUID `json:"employee_id"`
// 	EstimateHours string    `json:"estimate_hours"`
// 	Payment       bool      `json:"payment"`
// 	ProjectID     uuid.UUID `json:"project_id"`
// 	StatusHours   string    `json:"status_hours"`
// 	Month         string    `json:"month"`
// 	ProjectName   string    `json:"project_name"`
// 	ProjectStatus string    `json:"project_status"`
// 	WhatDid       string    `json:"what_did"`
// 	BH            float64   `json:"bh"`
// 	SH            float64   `json:"sh"`
// 	DH            float64   `json:"dh"`
// 	BHGS          string    `json:"bhgs"`
// 	WeekNumber    float64   `json:"week_number"`
// 	DayNumber     float64   `json:"day_number"`
// 	MonthNumber   float64   `json:"month_number"`
// 	PH            float64   `json:"ph"`
// 	ExpertiseID   uuid.UUID `json:"expertise_id"`
// 	Overtime      bool      `json:"overtime"`
// 	PCB           bool      `json:"pcb"`
// 	TaskEstimate  string    `json:"task_estimate"`
// 	PersonID      uuid.UUID `json:"person_id"`
// 	IDField       string    `json:"id_field"`
// 	ET            string    `json:"et"`
// 	Priority      string    `json:"priority"`
// 	MainTask      string    `json:"main_task"`
// 	TargetTask    string    `json:"target_task"`
// 	CR            bool      `json:"cr"`
// 	LastUpdate    time.Time `json:"last_update"`
// 	CreatedAt     time.Time `json:"created"`
// }

type timeDB struct {
	ID            uuid.UUID `db:"time_id"`
	TotalHours    float64   `db:"total_hours"`
	PayableHours  float64   `db:"payable_hours"`
	TaskID        uuid.UUID `db:"task_id"`
	Direction     string    `db:"direction"`
	WorkDate      time.Time `db:"work_date"`
	EmployeeID    uuid.UUID `db:"employee_id"`
	EstimateHours string    `db:"estimate_hours"`
	Payment       bool      `db:"payment"`
	ProjectID     uuid.UUID `db:"project_id"`
	StatusHours   string    `db:"status_hours"`
	Month         string    `db:"month"`
	ProjectStatus string    `db:"project_status"`
	WhatDid       string    `db:"what_did"`
	BH            float64   `db:"bh"`
	SH            float64   `db:"sh"`
	DH            float64   `db:"dh"`
	BHGS          string    `db:"bhgs"`
	WeekNumber    float64   `db:"week_number"`
	DayNumber     float64   `db:"day_number"`
	MonthNumber   float64   `db:"month_number"`
	PH            float64   `db:"ph"`
	ExpertiseID   uuid.UUID `db:"expertise_id"`
	Overtime      bool      `db:"overtime"`
	PCB           bool      `db:"pcb"`
	PersonID      uuid.UUID `db:"person_id"`
	IDField       string    `db:"id_field"`
	ET            string    `db:"et"`
	Priority      string    `db:"priority"`
	MainTask      string    `db:"main_task"`
	TargetTask    string    `db:"target_task"`
	CR            bool      `db:"cr"`
	LastUpdate    time.Time `db:"last_update"`
	CreatedAt     time.Time `db:"created"`

	TaskName     string  `db:"task_name"`
	ProjectName  string  `db:"project_name"`
	WhoDid       string  `db:"who_did"`
	Expertise    string  `db:"expertise"`
	TaskEstimate float64 `db:"task_estimate"`
}

func (t *timeDB) ToEntity() *entity_time.Time {
	return &entity_time.Time{
		ID:            t.ID,
		TotalHours:    t.TotalHours,
		PayableHours:  t.PayableHours,
		TaskID:        t.TaskID,
		Direction:     t.Direction,
		WorkDate:      t.WorkDate,
		EmployeeID:    t.EmployeeID,
		Payment:       t.Payment,
		ProjectID:     t.ProjectID,
		StatusHours:   t.StatusHours,
		Month:         t.Month,
		ProjectStatus: t.ProjectStatus,
		WhatDid:       t.WhatDid,
		BH:            t.BH,
		SH:            t.SH,
		DH:            t.DH,
		BHGS:          t.BHGS,
		WeekNumber:    t.WeekNumber,
		DayNumber:     t.DayNumber,
		MonthNumber:   t.MonthNumber,
		PH:            t.PH,
		ExpertiseID:   t.ExpertiseID,
		Overtime:      t.Overtime,
		PCB:           t.PCB,
		PersonID:      t.PersonID,
		IDField:       t.IDField,
		ET:            t.ET,
		Priority:      t.Priority,
		MainTask:      t.MainTask,
		TargetTask:    t.TargetTask,
		CR:            t.CR,
		LastUpdate:    t.LastUpdate,
		CreatedAt:     t.CreatedAt,

		ProjectName:  t.ProjectName,
		TaskName:     t.TaskName,
		WhoDid:       t.WhoDid,
		Expertise:    t.Expertise,
		TaskEstimate: t.TaskEstimate,
	}
}

func (r *TimePostgresRepository) SetTime(ctx context.Context, time *entity_time.Time) error {
	timeDB := &timeDB{
		ID:            time.ID,
		TotalHours:    time.TotalHours,
		PayableHours:  time.PayableHours,
		TaskID:        time.TaskID,
		Direction:     time.Direction,
		WorkDate:      time.WorkDate,
		EmployeeID:    time.EmployeeID,
		Payment:       time.Payment,
		ProjectID:     time.ProjectID,
		StatusHours:   time.StatusHours,
		Month:         time.Month,
		ProjectName:   time.ProjectName,
		ProjectStatus: time.ProjectStatus,
		WhatDid:       time.WhatDid,
		BH:            time.BH,
		SH:            time.SH,
		DH:            time.DH,
		BHGS:          time.BHGS,
		WeekNumber:    time.WeekNumber,
		DayNumber:     time.DayNumber,
		MonthNumber:   time.MonthNumber,
		PH:            time.PH,
		ExpertiseID:   time.ExpertiseID,
		Overtime:      time.Overtime,
		PCB:           time.PCB,
		TaskEstimate:  time.TaskEstimate,
		PersonID:      time.PersonID,
		IDField:       time.IDField,
		ET:            time.ET,
		Priority:      time.Priority,
		MainTask:      time.MainTask,
		TargetTask:    time.TargetTask,
		CR:            time.CR,
		LastUpdate:    time.LastUpdate,
		CreatedAt:     time.CreatedAt,
	}
	if _, err := r.DB().NamedExec(`
		INSERT INTO times (
			time_id, total_hours, payable_hours, task_id, direction, work_date, employee_id, payment, project_id, status_hours, month, project_status, what_did, bh, sh, dh, bhgs, week_number, day_number, month_number, ph, expertise_id, overtime, pcb, person_id, id_field, et, priority, main_task, target_task, cr, last_update, created
		) VALUES (
			:time_id, :total_hours, :payable_hours, :task_id, :direction, :work_date, :employee_id, :payment, :project_id, :status_hours, :month, :project_status, :what_did, :bh, :sh, :dh, :bhgs, :week_number, :day_number, :month_number, :ph, :expertise_id, :overtime, :pcb, :person_id, :id_field, :et, :priority, :main_task, :target_task, :cr, :last_update, :created
		) ON CONFLICT (time_id) DO UPDATE SET
			total_hours = :total_hours,
			payable_hours = :payable_hours,
			task_id = :task_id,
			direction = :direction,
			work_date = :work_date,
			employee_id = :employee_id,
			payment = :payment,
			project_id = :project_id,
			status_hours = :status_hours,
			month = :month,
			project_status = :project_status,
			what_did = :what_did,
			bh = :bh,
			sh = :sh,
			dh = :dh,
			bhgs = :bhgs,
			week_number = :week_number,
			day_number = :day_number,
			month_number = :month_number,
			ph = :ph,
			expertise_id = :expertise_id,
			overtime = :overtime,
			pcb = :pcb,
			person_id = :person_id,
			id_field = :id_field,
			et = :et,
			priority = :priority,
			main_task = :main_task,
			target_task = :target_task,
			cr = :cr,
			last_update = :last_update,
			created = :created
	`, timeDB); err != nil {
		slog.Error("Error while setting time", "error", err)
		return err
	}
	return nil
}

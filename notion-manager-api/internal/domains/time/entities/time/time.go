package time

import (
	"time"

	"github.com/google/uuid"
)

type Time struct {
	ID            uuid.UUID `json:"id"`
	TotalHours    float64   `json:"total_hours"`
	PayableHours  float64   `json:"payable_hours"`
	TaskID        uuid.UUID `json:"task_id"`
	Direction     string    `json:"direction"`
	WorkDate      time.Time `json:"work_date"`
	EmployeeID    uuid.UUID `json:"employee_id"`
	Payment       bool      `json:"payment"`
	ProjectID     uuid.UUID `json:"project_id"`
	StatusHours   string    `json:"status_hours"`
	Month         string    `json:"month"`
	ProjectName   string    `json:"project_name"`
	ProjectStatus string    `json:"project_status"`
	WhatDid       string    `json:"what_did"`
	BH            float64   `json:"bh"`
	SH            float64   `json:"sh"`
	DH            float64   `json:"dh"`
	BHGS          string    `json:"bhgs"`
	WeekNumber    float64   `json:"week_number"`
	DayNumber     float64   `json:"day_number"`
	MonthNumber   float64   `json:"month_number"`
	PH            float64   `json:"ph"`
	ExpertiseID   uuid.UUID `json:"expertise_id"`
	Overtime      bool      `json:"overtime"`
	PCB           bool      `json:"pcb"`
	PersonID      uuid.UUID `json:"person_id"`
	IDField       string    `json:"id_field"`
	ET            string    `json:"et"`
	Priority      string    `json:"priority"`
	MainTask      string    `json:"main_task"`
	TargetTask    string    `json:"target_task"`
	CR            bool      `json:"cr"`
	LastUpdate    time.Time `json:"last_update"`
	CreatedAt     time.Time `json:"created"`

	WhoDid       string  `json:"who_did"`
	Expertise    string  `json:"expertise"`
	TaskName     string  `json:"task_name"`
	TaskEstimate float64 `json:"task_estimate"`
}

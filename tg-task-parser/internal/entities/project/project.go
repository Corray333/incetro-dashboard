package project

import "github.com/google/uuid"

type ProjectStatus string

const (
	StatusActive ProjectStatus = "Активный"
)

type Project struct {
	ID   uuid.UUID `json:"id" db:"project_id"`
	Name string    `json:"name" db:"name"`
}

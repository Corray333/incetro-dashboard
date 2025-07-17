package client

import (
	"time"

	"github.com/google/uuid"
)

type projectI interface {
	GetName() string
}

type Client struct {
	ID         uuid.UUID   `json:"id"`
	Name       string      `json:"name"`
	Status     Status      `json:"status"`
	Source     string      `json:"source"`
	UniqueID   int64       `json:"unique_id"`
	CreatedAt  time.Time   `json:"created_at"`
	UpdatedAt  time.Time   `json:"updated_at"`
	ProjectIDs []uuid.UUID `json:"project_ids"`
	Projects   []projectI  `json:"projects"`
}

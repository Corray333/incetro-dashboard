package task

import "github.com/google/uuid"

type Filter struct {
	ProjectID uuid.UUID `json:"projectID"`
}

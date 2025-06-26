package employee

import "github.com/google/uuid"

type Filter struct {
	ID        uuid.UUID `json:"id"`
	ProfileID uuid.UUID `json:"profile_id"`
}

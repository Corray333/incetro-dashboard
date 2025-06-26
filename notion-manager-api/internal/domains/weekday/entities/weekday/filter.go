package weekday

import (
	"time"

	"github.com/google/uuid"
)

type Filter struct {
	ID            uuid.UUID `json:"id"`
	UpdatedAtFrom time.Time `json:"updated_at_from"`
	UpdatedAtTo   time.Time `json:"updated_at_to"`
	Notified      *bool     `json:"notified"`
}

package feedback

import (
	"time"

	"github.com/google/uuid"
)

type Feedback struct {
	ID          uuid.UUID `json:"id"`
	Text        string    `json:"text"`
	Type        string    `json:"type"`
	Priority    string    `json:"priority"`
	TaskID      uuid.UUID `json:"task"`
	ProjectID   uuid.UUID `json:"project"`
	CreatedDate time.Time `json:"createdDate"`
	Direction   string    `json:"direction"`
	Status      string    `json:"status"`
}

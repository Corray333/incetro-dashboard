package employee

import "github.com/google/uuid"

type Employee struct {
	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
}

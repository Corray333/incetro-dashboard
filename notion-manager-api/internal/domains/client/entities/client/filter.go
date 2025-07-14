package client

import "github.com/google/uuid"

type Filter struct {
	IDs    []uuid.UUID `json:"ids"`
	Status *Status     `json:"status"`
	Source *string     `json:"source"`
	Name   *string     `json:"name"`
}
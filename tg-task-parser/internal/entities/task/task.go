package task

import "encoding/json"

// Tag and Mention types mirror message entity concepts to avoid cross-package import cycles
// and allow the Task entity to carry parsed data needed for creation flows.
type Tag string

type Mention string

type Task struct {
	Title     string          `json:"title"`
	Link      string          `json:"link"`
	Body      json.RawMessage `json:"body"`
	PlainBody string          `json:"plainBody"`
	Hashtags  []Tag           `json:"hashtags"`
	Executors []Mention       `json:"executors"`
	Assignee  Mention         `json:"assignee"`
	Images    []string        `json:"images"`
}

package task

// Tag and Mention types mirror message entity concepts to avoid cross-package import cycles
// and allow the Task entity to carry parsed data needed for creation flows.
type Tag string

type Mention string

type Task struct {
	Text     string    `json:"text"`
	Link     string    `json:"link"`
	Hashtags []Tag     `json:"hashtags"`
	Mentions []Mention `json:"mentions"`
}

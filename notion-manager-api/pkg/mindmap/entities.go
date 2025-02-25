package mindmap

type Task struct {
	Title    string  `json:"title"`
	Link     string  `json:"link"`
	Hours    float64 `json:"hours"`
	Subtasks []Task  `json:"subtasks"`
}

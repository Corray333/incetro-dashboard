package client

type Status string

const (
	StatusActive Status = "Active"
	StatusPaused Status = "Paused"
	StatusDone   Status = "Done"
)
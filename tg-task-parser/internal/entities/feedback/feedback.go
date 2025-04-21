package feedback

import (
	"time"

	"github.com/google/uuid"
)

type Status string

const (
	StatusDone             Status = "Готова"
	StatusInProgress       Status = "В работе"
	StatusWaitingForClient Status = "Ожидает ответа клиента"
	StatusClientCheck      Status = "Проверка клиентом"
	StatusEstimateAgree    Status = "Согласование оценки"
	StatusNoEstimate       Status = "Нет оценки"
)

var NotStartedStatuses = []Status{
	StatusNoEstimate,
}

var ActiveStatuses = []Status{
	StatusInProgress,
	StatusWaitingForClient,
	StatusClientCheck,
	StatusEstimateAgree,
}

var DoneStatuses = []Status{
	StatusDone,
}

type Feedback struct {
	ID          uuid.UUID `json:"id"`
	Text        string    `json:"text"`
	Type        string    `json:"type"`
	Priority    string    `json:"priority"`
	TaskID      uuid.UUID `json:"task"`
	ProjectID   uuid.UUID `json:"project"`
	CreatedDate time.Time `json:"createdDate"`
	Direction   string    `json:"direction"`
	Status      Status    `json:"status"`
}

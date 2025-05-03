package task

type Status string

const (
	StatusForming       Status = "Формируется"
	StatusCanDo         Status = "Можно делать"
	StatusOnHold        Status = "На паузе"
	StatusWaiting       Status = "Ожидание"
	StatusInProgress    Status = "В работе"
	StatusNeedDiscuss   Status = "Надо обсудить"
	StatusCodeReview    Status = "Код-ревью"
	StatusInternalCheck Status = "Внутренняя проверка"
	StatusReadyToUpload Status = "Можно выгружать"
	StatusClientCheck   Status = "Проверка клиентом"
	StatusCancelled     Status = "Отменена"
	StatusDone          Status = "Готова"
)

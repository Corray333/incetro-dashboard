package employee

import "github.com/google/uuid"

type Employee struct {
	ID         uuid.UUID `db:"id"`
	TgID       int64     `db:"tg_id"`
	TgUsername string    `db:"tg_username"`
	FIO        string    `db:"fio"`
}

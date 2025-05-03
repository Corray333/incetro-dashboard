-- +goose Up
-- +goose StatementBegin
-- type taskOutboxMsgDB struct {
-- 	ID         int64     `db:"task_msg_id"`
-- 	Task       string    `db:"task"`
-- 	Estimate   float64   `db:"estimate"`
-- 	Priority   string    `db:"priority"`
-- 	Start      time.Time `db:"start"`
-- 	End        time.Time `db:"end"`
-- 	ExecutorID uuid.UUID `db:"executor_id"`
-- 	ProjectID  uuid.UUID `db:"project_id"`
-- }
CREATE TABLE task_outbox(
    task_msg_id SERIAL NOT NULL,
    task TEXT NOT NULL,
    estimate FLOAT NOT NULL DEFAULT 0,
    priority TEXT NOT NULL DEFAULT '',
    deadline_start TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deadline_end TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    executor_id UUID NOT NULL,
    project_id UUID NOT NULL,
    CONSTRAINT task_outbox_pkey PRIMARY KEY (task_msg_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE task_outbox;
-- +goose StatementEnd

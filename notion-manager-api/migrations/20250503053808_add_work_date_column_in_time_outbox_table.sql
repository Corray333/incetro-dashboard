-- +goose Up
-- +goose StatementBegin
ALTER TABLE time_outbox ADD COLUMN work_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE time_outbox DROP COLUMN work_date;
-- +goose StatementEnd

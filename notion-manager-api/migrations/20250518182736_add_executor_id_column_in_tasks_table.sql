-- +goose Up
-- +goose StatementBegin
ALTER TABLE tasks ADD COLUMN executor_id UUID;
ALTER TABLE tasks ADD COLUMN responsible_id UUID;
ALTER TABLE tasks ADD COLUMN main_task TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tasks DROP COLUMN executor_id;
ALTER TABLE tasks DROP COLUMN responsible_id;
ALTER TABLE tasks DROP COLUMN main_task;
-- +goose StatementEnd

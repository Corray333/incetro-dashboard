-- +goose Up
-- +goose StatementBegin
ALTER TABLE task_tag
ALTER COLUMN task_id TYPE uuid USING task_id::uuid;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tasks
ALTER COLUMN task_tag TYPE text USING task_id::text;
-- +goose StatementEnd

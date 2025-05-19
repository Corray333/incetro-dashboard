-- +goose Up
-- +goose StatementBegin
ALTER TABLE tasks
ALTER COLUMN task_id TYPE uuid USING task_id::uuid;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE your_table
ALTER COLUMN tasks TYPE text USING tasks::text;
-- +goose StatementEnd

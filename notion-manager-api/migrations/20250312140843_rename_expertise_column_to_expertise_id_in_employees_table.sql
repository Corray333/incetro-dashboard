-- +goose Up
-- +goose StatementBegin
ALTER TABLE employees RENAME COLUMN expertise TO expertise_id;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE employees RENAME COLUMN expertise_id TO expertise;
-- +goose StatementEnd

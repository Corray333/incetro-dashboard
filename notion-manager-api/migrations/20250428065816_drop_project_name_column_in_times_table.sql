-- +goose Up
-- +goose StatementBegin
ALTER TABLE times DROP COLUMN project_name;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE times ADD COLUMN project_name TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

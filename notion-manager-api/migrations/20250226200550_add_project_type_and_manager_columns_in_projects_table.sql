-- +goose Up
-- +goose StatementBegin
ALTER TABLE projects ADD COLUMN type TEXT NOT NULL DEFAULT '';
ALTER TABLE projects ADD COLUMN manager_id TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE projects DROP COLUMN type;
ALTER TABLE projects DROP COLUMN manager_id;
-- +goose StatementEnd

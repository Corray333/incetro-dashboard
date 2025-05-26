-- +goose Up
-- +goose StatementBegin
ALTER TABLE projects ADD COLUMN sheets_link TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE projects DROP COLUMN sheets_link;
-- +goose StatementEnd

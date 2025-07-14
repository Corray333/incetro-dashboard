-- +goose Up
-- +goose StatementBegin
ALTER TABLE projects ADD COLUMN client_id UUID;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE projects DROP COLUMN client_id;
-- +goose StatementEnd
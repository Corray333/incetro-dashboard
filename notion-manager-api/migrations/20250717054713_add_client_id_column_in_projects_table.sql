-- +goose Up
-- +goose StatementBegin
ALTER TABLE projects ADD COLUMN client_id UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE projects DROP COLUMN client_id;
-- +goose StatementEnd

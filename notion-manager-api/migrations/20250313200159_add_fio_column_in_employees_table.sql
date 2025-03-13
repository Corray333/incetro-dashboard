-- +goose Up
-- +goose StatementBegin
ALTER TABLE employees ADD COLUMN fio TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE employees DROP COLUMN fio;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
ALTER TABLE employees
ALTER COLUMN employee_id TYPE uuid USING employee_id::uuid;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE employees
ALTER COLUMN employee_id TYPE integer USING employee_id::integer;
-- +goose StatementEnd

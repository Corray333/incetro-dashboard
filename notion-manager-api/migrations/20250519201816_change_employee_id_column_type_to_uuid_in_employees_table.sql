-- +goose Up
-- +goose StatementBegin
ALTER TABLE employees
ALTER COLUMN employee_id TYPE uuid USING employee_id::uuid;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd

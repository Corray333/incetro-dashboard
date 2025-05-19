-- +goose Up
-- +goose StatementBegin
ALTER TABLE expertise
ALTER COLUMN expertise_id TYPE uuid USING expertise_id::uuid;


ALTER TABLE employees
ALTER COLUMN expertise_id DROP DEFAULT;
ALTER TABLE employees
ALTER COLUMN expertise_id DROP NOT NULL;
UPDATE employees
SET expertise_id = NULL
WHERE expertise_id = '';
ALTER TABLE employees
ALTER COLUMN expertise_id TYPE uuid USING expertise_id::uuid;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE expertise
ALTER COLUMN expertise_id TYPE integer USING expertise_id::text;
ALTER TABLE employees
ALTER COLUMN expertise_id TYPE integer USING expertise_id::text;
ALTER TABLE employees
ALTER COLUMN expertise_id SET DEFAULT '';

-- +goose StatementEnd

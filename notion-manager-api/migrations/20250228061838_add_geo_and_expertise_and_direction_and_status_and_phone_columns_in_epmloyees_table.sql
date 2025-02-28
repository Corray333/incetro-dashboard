-- +goose Up
-- +goose StatementBegin
ALTER TABLE employees ADD COLUMN geo TEXT NOT NULL DEFAULT '';
ALTER TABLE employees ADD COLUMN expertise TEXT NOT NULL DEFAULT '';
ALTER TABLE employees ADD COLUMN direction TEXT NOT NULL DEFAULT '';
ALTER TABLE employees ADD COLUMN status TEXT NOT NULL DEFAULT '';
ALTER TABLE employees ADD COLUMN phone TEXT NOT NULL DEFAULT '';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE employees DROP COLUMN geo;
ALTER TABLE employees DROP COLUMN expertise;
ALTER TABLE employees DROP COLUMN direction;
ALTER TABLE employees DROP COLUMN status;
ALTER TABLE employees DROP COLUMN phone;
-- +goose StatementEnd

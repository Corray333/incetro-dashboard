-- +goose Up
-- +goose StatementBegin
ALTER TABLE times DROP COLUMN estimate_hours;
ALTER TABLE times DROP COLUMN task_estimate;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE times ADD COLUMN estimate_hours FLOAT NOT NULL DEFAULT 0;
ALTER TABLE times ADD COLUMN task_estimate FLOAT NOT NULL DEFAULT 0;
-- +goose StatementEnd

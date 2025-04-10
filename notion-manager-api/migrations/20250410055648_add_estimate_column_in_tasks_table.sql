-- +goose Up
-- +goose StatementBegin
ALTER TABLE tasks ADD COLUMN estimate FLOAT NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tasks DROP COLUMN estimate;
-- +goose StatementEnd

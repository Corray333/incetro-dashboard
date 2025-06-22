-- +goose Up
-- +goose StatementBegin
ALTER TABLE employees ADD COLUMN unique_id BIGINT NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE employees DROP COLUMN unique_id;
-- +goose StatementEnd

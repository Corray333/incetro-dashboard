-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS employees (
    employee_id VARCHAR(36) PRIMARY KEY,
    username TEXT NOT NULL DEFAULT '',
    icon TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL DEFAULT '',
    profile_id VARCHAR(36) NOT NULL DEFAULT '',
    tg_username TEXT NOT NULL DEFAULT '',
    tg_id BIGINT NOT NULL DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS employees;
-- +goose StatementEnd

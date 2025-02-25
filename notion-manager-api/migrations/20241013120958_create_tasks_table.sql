-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tasks(
    task_id VARCHAR(36) PRIMARY KEY,
    title TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT '',
    employee_id VARCHAR(36),
    project_id VARCHAR(36),
    start_time BIGINT NOT NULL DEFAULT 0,
    end_time BIGINT NOT NULL DEFAULT 0
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS tasks;
-- +goose StatementEnd

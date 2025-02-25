-- +goose Up
-- +goose StatementBegin
CREATE TABLE task_tag (
    task_id VARCHAR(36),
    tag TEXT,
    PRIMARY KEY (task_id, tag)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE task_tag;
-- +goose StatementEnd

-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS invalid_rows(
    id TEXT NOT NULL,
    description TEXT NOT NULL,
    employee VARCHAR(36) NOT NULL,
    employee_id VARCHAR(36) NOT NULL,
    CONSTRAINT invalid_rows_pkey PRIMARY KEY (id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE invalid_rows;
-- +goose StatementEnd

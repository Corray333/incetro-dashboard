-- +goose Up
-- +goose StatementBegin
CREATE TABLE expertise (
    expertise_id VARCHAR(36) NOT NULL,
    name TEXT NOT NULL,
    direction TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (expertise_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE expertise;
-- +goose StatementEnd

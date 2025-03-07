-- +goose Up
-- +goose StatementBegin
CREATE TABLE expertise (
    name TEXT NOT NULL,
    direction TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (name)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE expertise;
-- +goose StatementEnd

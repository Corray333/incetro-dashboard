-- +goose Up
-- +goose StatementBegin
CREATE TABLE message_meta(
    chat_id BIGINT NOT NULL,
    message_id BIGINT NOT NULL,
    meta JSONB NOT NULL,
    PRIMARY KEY (chat_id, message_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE message_meta;
-- +goose StatementEnd

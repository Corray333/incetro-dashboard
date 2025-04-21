-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS chats (
    chat_id BIGINT NOT NULL,
    project_id UUID NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE chats;
-- +goose StatementEnd

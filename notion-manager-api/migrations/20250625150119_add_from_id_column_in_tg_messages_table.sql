-- +goose Up
-- +goose StatementBegin
ALTER TABLE tg_messages ADD COLUMN from_id BIGINT NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE tg_messages DROP COLUMN from_id;
-- +goose StatementEnd

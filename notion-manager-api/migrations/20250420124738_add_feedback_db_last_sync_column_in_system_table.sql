-- +goose Up
-- +goose StatementBegin
ALTER TABLE system
ADD COLUMN feedback_db_last_sync TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT '1970-01-01 00:00:00';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE system
DROP COLUMN feedback_db_last_sync;
-- +goose StatementEnd

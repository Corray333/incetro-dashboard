-- +goose Up
-- +goose StatementBegin
ALTER TABLE "system" DROP COLUMN times_db_last_sync;
ALTER TABLE "system" ADD COLUMN times_db_last_sync TIMESTAMPTZ NOT NULL DEFAULT '1970-01-01 00:00:00';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "system" DROP COLUMN times_db_last_sync;
ALTER TABLE "system" ADD COLUMN times_db_last_sync BIGINT NOT NULL DEFAULT 0;
-- +goose StatementEnd

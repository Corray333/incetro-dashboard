-- +goose Up
-- +goose StatementBegin
ALTER TABLE system ADD COLUMN clients_db_last_sync TIMESTAMPTZ NOT NULL DEFAULT '1970-01-01 00:00:00+00';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE system DROP COLUMN clients_db_last_sync;
-- +goose StatementEnd
-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS clients (
    client_id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    status TEXT NOT NULL,
    source TEXT,
    unique_id TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    project_ids UUID[] DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_clients_status ON clients(status);
CREATE INDEX IF NOT EXISTS idx_clients_source ON clients(source);
CREATE INDEX IF NOT EXISTS idx_clients_name ON clients(name);
CREATE INDEX IF NOT EXISTS idx_clients_unique_id ON clients(unique_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clients;
-- +goose StatementEnd
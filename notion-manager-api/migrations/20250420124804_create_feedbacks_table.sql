-- +goose Up
-- +goose StatementBegin
CREATE TABLE feedbacks (
    feedback_id UUID PRIMARY KEY,
    text TEXT NOT NULL,
    type TEXT NOT NULL,
    priority TEXT NOT NULL,
    task_id UUID NOT NULL,
    project_id UUID NOT NULL,
    created_date TIMESTAMPTZ NOT NULL,
    direction TEXT NOT NULL,
    status TEXT NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE feedbacks;
-- +goose StatementEnd

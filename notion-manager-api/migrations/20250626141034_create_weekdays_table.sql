-- +goose Up
-- +goose StatementBegin
CREATE TABLE weekdays (
    weekday_id UUID NOT NULL,
    employee_id UUID NOT NULL,
    category VARCHAR(255) NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    reason TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    notified BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (weekday_id)
);
UPDATE weekdays SET notified = TRUE;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE weekdays;
-- +goose StatementEnd

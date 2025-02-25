-- +goose Up
-- +goose StatementBegin
CREATE TABLE employee_notification_flag (
    employee_id VARCHAR(36),
    flag TEXT,
    PRIMARY KEY (employee_id, flag)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE employee_notification_flag;
-- +goose StatementEnd

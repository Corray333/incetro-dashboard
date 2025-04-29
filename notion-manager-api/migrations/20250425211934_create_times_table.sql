-- +goose Up
-- +goose StatementBegin
CREATE TABLE times (
    time_id         UUID PRIMARY KEY NOT NULL,
    total_hours     DOUBLE PRECISION NOT NULL DEFAULT 0,
    payable_hours   DOUBLE PRECISION NOT NULL DEFAULT 0,
    task_id         UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
    direction       TEXT NOT NULL DEFAULT '',
    work_date       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    employee_id     UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
    estimate_hours  TEXT NOT NULL DEFAULT '',
    payment         BOOLEAN NOT NULL DEFAULT FALSE,
    project_id      UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
    status_hours    TEXT NOT NULL DEFAULT '',
    month           TEXT NOT NULL DEFAULT '',
    project_name    TEXT NOT NULL DEFAULT '',
    project_status  TEXT NOT NULL DEFAULT '',
    what_did        TEXT NOT NULL DEFAULT '',
    bh              DOUBLE PRECISION NOT NULL DEFAULT 0,
    sh              DOUBLE PRECISION NOT NULL DEFAULT 0,
    dh              DOUBLE PRECISION NOT NULL DEFAULT 0,
    bhgs            TEXT NOT NULL DEFAULT '',
    week_number     DOUBLE PRECISION NOT NULL DEFAULT 0,
    day_number      DOUBLE PRECISION NOT NULL DEFAULT 0,
    month_number    DOUBLE PRECISION NOT NULL DEFAULT 0,
    ph              DOUBLE PRECISION NOT NULL DEFAULT 0,
    expertise_id    UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
    overtime        BOOLEAN NOT NULL DEFAULT FALSE,
    pcb             BOOLEAN NOT NULL DEFAULT FALSE,
    task_estimate   TEXT NOT NULL DEFAULT '',
    person_id       UUID NOT NULL DEFAULT '00000000-0000-0000-0000-000000000000',
    id_field        TEXT NOT NULL DEFAULT '',
    et              TEXT NOT NULL DEFAULT '',
    priority        TEXT NOT NULL DEFAULT '',
    main_task       TEXT NOT NULL DEFAULT '',
    target_task     TEXT NOT NULL DEFAULT '',
    cr              BOOLEAN NOT NULL DEFAULT FALSE,
    last_update     TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created         TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS times;
-- +goose StatementEnd

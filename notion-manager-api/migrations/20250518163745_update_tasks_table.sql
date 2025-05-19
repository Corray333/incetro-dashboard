-- +goose Up
-- +goose StatementBegin
-- 2. Добавляем обязательные поля времён создания и обновления:
ALTER TABLE tasks
    ADD COLUMN created_time    timestamp without time zone NOT NULL DEFAULT now(),
    ADD COLUMN last_edited_time timestamp without time zone NOT NULL DEFAULT now();

-- 3. Добавляем прочие новые текстовые/числовые поля:
ALTER TABLE tasks
    ADD COLUMN priority        text NOT NULL DEFAULT ''::text,
    ADD COLUMN parent_id       UUID,
    ADD COLUMN creator_id      UUID,
    ADD COLUMN ikp             text;

-- 4. Добавляем массивы идентификаторов:
-- ALTER TABLE tasks
--     ADD COLUMN responsible_ids  varchar(36)[] NOT NULL DEFAULT ARRAY[]::varchar(36)[],
--     ADD COLUMN executor_ids     varchar(36)[] NOT NULL DEFAULT ARRAY[]::varchar(36)[],
--     ADD COLUMN tags             text[]        NOT NULL DEFAULT ARRAY[]::text[],
--     ADD COLUMN subtasks_ids     varchar(36)[] NOT NULL DEFAULT ARRAY[]::varchar(36)[];

-- 5. Добавляем поля для “ссылок”:
ALTER TABLE tasks
    ADD COLUMN previous_id     UUID,
    ADD COLUMN next_id         UUID;

-- 6. Добавляем числовые метрики:
ALTER TABLE tasks
    ADD COLUMN total_hours      double precision NOT NULL DEFAULT 0,
    ADD COLUMN tbh              double precision NOT NULL DEFAULT 0,
    ADD COLUMN cp               double precision NOT NULL DEFAULT 0,
    ADD COLUMN total_estimate   double precision NOT NULL DEFAULT 0,
    ADD COLUMN plan_fact        double precision NOT NULL DEFAULT 0,
    ADD COLUMN duration         double precision NOT NULL DEFAULT 0,
    ADD COLUMN cr               double precision NOT NULL DEFAULT 0;

-- 7. Обработка времени начала и окончания:
--    Сначала создаём новые колонки типа timestamp,
--    затем заполняем их из старых bigint (UNIX epoch),
--    и только после этого удаляем старые.
ALTER TABLE tasks
    ADD COLUMN start           timestamp without time zone,
    ADD COLUMN "end"             timestamp without time zone;

UPDATE tasks
SET
    start = to_timestamp(start_time),
    "end" = to_timestamp(end_time);

ALTER TABLE tasks
    DROP COLUMN start_time,
    DROP COLUMN end_time;

-- 8. Переименование колонки status (если нужно оставить без изменений, этот шаг можно пропустить):
--    здесь оставляем как есть, т.к. совпадает с Go-тегом `status`

-- 9. Удаляем устаревшее поле employee_id:
ALTER TABLE tasks
    DROP COLUMN IF EXISTS employee_id;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- 1. Восстанавливаем удалённые колонки и ограничения:
ALTER TABLE tasks
    DROP CONSTRAINT IF EXISTS tasks_pkey,
    ADD COLUMN employee_id UUID,  -- восстанавливаем удалённую колонку
    ADD COLUMN start_time  bigint,
    ADD COLUMN end_time    bigint;

-- 2. Конвертируем timestamp обратно в bigint (UNIX epoch):
UPDATE tasks
SET
    start_time = EXTRACT(EPOCH FROM start),
    end_time   = EXTRACT(EPOCH FROM "end");

ALTER TABLE tasks
    DROP COLUMN start,
    DROP COLUMN "end";

-- 3. Удаляем новые колонки:
ALTER TABLE tasks
    DROP COLUMN created_time,
    DROP COLUMN last_edited_time,
    DROP COLUMN priority,
    DROP COLUMN parent_id,
    DROP COLUMN creator_id,
    DROP COLUMN ikp,
    DROP COLUMN previous_id,
    DROP COLUMN next_id,
    DROP COLUMN total_hours,
    DROP COLUMN tbh,
    DROP COLUMN cp,
    DROP COLUMN total_estimate,
    DROP COLUMN plan_fact,
    DROP COLUMN duration,
    DROP COLUMN cr;

-- 4. Восстанавливаем первичный ключ, если требуется:
ALTER TABLE tasks
    ADD PRIMARY KEY (id);
    
-- +goose StatementEnd

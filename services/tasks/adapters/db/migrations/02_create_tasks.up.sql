CREATE TABLE IF NOT EXISTS tasks (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    category_id BIGINT NULL REFERENCES categories(id) ON DELETE SET NULL,

    name        text NOT NULL,
    description text NULL,

    -- 0 todo, 1 in_progress, 2 done, 3 archived
    status      smallint NOT NULL DEFAULT 0,

    created_at  timestamptz NOT NULL DEFAULT now(),
    updated_at  timestamptz NOT NULL DEFAULT now()
);

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1
    FROM pg_constraint
    WHERE conname = 'chk_tasks_status'
      AND conrelid = 'tasks'::regclass
  ) THEN
ALTER TABLE tasks
    ADD CONSTRAINT chk_tasks_status
        CHECK (status IN (0, 1, 2, 3));
END IF;
END$$;


CREATE INDEX IF NOT EXISTS idx_tasks_category_id
    ON tasks (category_id);

CREATE INDEX IF NOT EXISTS idx_tasks_status
    ON tasks (status);

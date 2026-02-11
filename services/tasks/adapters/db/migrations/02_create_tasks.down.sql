DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_category_id;
ALTER TABLE tasks DROP CONSTRAINT IF EXISTS chk_tasks_status;
DROP TABLE IF EXISTS tasks;

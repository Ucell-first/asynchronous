-- Indexlarni o'chirish
DROP INDEX IF EXISTS idx_task_results_task_id;
DROP INDEX IF EXISTS idx_tasks_status;
DROP INDEX IF EXISTS idx_tasks_user_id;
DROP INDEX IF EXISTS idx_tasks_creator_id;

-- Jadvallarni o'chirish
DROP TABLE IF EXISTS task_results;
DROP TABLE IF EXISTS tasks;
DROP TABLE IF EXISTS users;
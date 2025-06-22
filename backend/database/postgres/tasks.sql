-- name: TaskCreateForUserId :one
INSERT INTO tasks (user_id, category_id, type, url)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: TaskListByUserId :many
SELECT id, user_id, upload_id, category_id, type, url, status, failed_reason, token_count, created_at, updated_at
FROM tasks
WHERE user_id = $1
  AND (sqlc.narg('upload_id')::BIGINT IS NULL OR upload_id = sqlc.narg('upload_id')::BIGINT)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: TaskCountByUserId :one
SELECT COUNT(*)
FROM tasks
WHERE user_id = $1
  AND (sqlc.narg('upload_id')::BIGINT IS NULL OR upload_id = sqlc.narg('upload_id')::BIGINT);

-- name: TaskOverviewByUserId :many
WITH daily_stats AS (
    SELECT 
        d.day_date,
        COUNT(tasks.id) FILTER (WHERE DATE(tasks.created_at) = d.day_date) as submitted,
        COUNT(tasks.id) FILTER (WHERE tasks.status IN ('queuing', 'processing') AND DATE(tasks.created_at) = d.day_date) as pending,
        COUNT(tasks.id) FILTER (WHERE tasks.status = 'completed' AND DATE(tasks.created_at) = d.day_date) as completed,
        COUNT(tasks.id) FILTER (WHERE tasks.status = 'failed' AND DATE(tasks.created_at) = d.day_date) as failed
    FROM (
        SELECT DATE(NOW() - INTERVAL '1 day' * generate_series(0, 6)) as day_date
    ) d
    LEFT JOIN tasks ON DATE(tasks.created_at) = d.day_date AND tasks.user_id = $1
    GROUP BY d.day_date
    ORDER BY d.day_date DESC
),
token_stats AS (
    SELECT 
        COALESCE(SUM(token_count) FILTER (WHERE status = 'completed' AND updated_at >= NOW() - INTERVAL '7 days'), 0) as token_histories,
        COALESCE(SUM(token_count) FILTER (WHERE status = 'completed'), 0) as token_count
    FROM tasks 
    WHERE user_id = $1
),
pool_stats AS (
    SELECT COALESCE(SUM(token_count) FILTER (WHERE status = 'completed'), 0) as pool_token_count
    FROM tasks
)
SELECT
    daily_stats.submitted,
    daily_stats.pending,
    daily_stats.completed,
    daily_stats.failed,
    token_stats.token_histories::INTEGER as token_histories,
    token_stats.token_count::INTEGER as token_count,
    pool_stats.pool_token_count::INTEGER as pool_token_count
FROM daily_stats, token_stats, pool_stats;
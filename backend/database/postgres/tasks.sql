-- name: TaskCreateForUserId :one
INSERT INTO tasks (user_id, upload_id, category_id, type, source, is_raw, title, content)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: TaskListByUserId :many
SELECT id, user_id, upload_id, category_id, type, source, status, failed_reason, token_count, created_at, updated_at
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

-- name: TaskGetById :one
SELECT sqlc.embed(tasks), sqlc.embed(users), sqlc.embed(categories)
FROM tasks
LEFT JOIN users ON tasks.user_id = users.id
LEFT JOIN categories ON tasks.category_id = categories.id
WHERE tasks.id = $1;

-- name: TaskOverviewByUserId :many
WITH token_stats AS (
    SELECT
        COALESCE(SUM(token_count) FILTER (WHERE status = 'completed'), 0) as token_count
    FROM tasks
    WHERE tasks.user_id = $1
),
daily_stats AS (
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
    ORDER BY d.day_date
)
SELECT
    token_stats.token_count::INTEGER as token_count,
    daily_stats.submitted,
    daily_stats.pending,
    daily_stats.completed,
    daily_stats.failed
FROM daily_stats, token_stats;

-- name: PoolTokenOverviewByCategory :many
SELECT
    categories.id as category_id,
    categories.name as category_name,
    COALESCE(SUM(tasks.token_count) FILTER (WHERE tasks.status = 'completed'), 0)::INTEGER as token_count
FROM categories
LEFT JOIN tasks ON categories.id = tasks.category_id
GROUP BY categories.id, categories.name
ORDER BY categories.name;

-- name: TaskClaimPending :one
UPDATE tasks
SET status = 'processing'
WHERE id = (
    SELECT t.id
    FROM tasks t
    WHERE t.status = 'queuing'
      AND t.user_id = (
        SELECT user_id
        FROM (SELECT DISTINCT user_id FROM tasks WHERE status = 'queuing') users
        ORDER BY RANDOM()
        LIMIT 1
    )
    ORDER BY t.created_at
    LIMIT 1
    FOR UPDATE SKIP LOCKED
)
RETURNING *;

-- name: TaskUpdateCompleted :exec
UPDATE tasks
SET status          = 'completed',
    title           = COALESCE($2, title),
    content         = COALESCE($3, content),
    token_count     = COALESCE($4, token_count),
    revised_task_id = COALESCE($5, revised_task_id)
WHERE id = $1;

-- name: TaskUpdateFailed :exec
UPDATE tasks
SET status = 'failed',
    failed_reason = $2,
    title = COALESCE($3, title),
    content = COALESCE($4, content),
    token_count = COALESCE($5, token_count)
WHERE id = $1;

-- name: TaskResetFailed :many
UPDATE tasks
SET status = 'queuing',
    failed_reason = NULL,
    title = NULL,
    content = NULL,
    token_count = 0
WHERE status = 'failed'
RETURNING *;

-- name: TaskListCompleted :many
SELECT *
FROM tasks
WHERE status = 'completed'
ORDER BY created_at;

-- name: TaskTotalCompletedByUserId :one
SELECT COUNT(*) as total_completed
FROM tasks
WHERE user_id = $1 AND status = 'completed';

-- name: TaskTotalFailedByUserId :one
SELECT COUNT(*) as total_failed
FROM tasks
WHERE user_id = $1 AND status = 'failed';

-- name: TaskTotalPendingByUserId :one
SELECT COUNT(*) as total_pending
FROM tasks
WHERE user_id = $1 AND status IN ('queuing', 'processing');

-- name: TaskListFailedRawDuplicates :many
SELECT *
FROM tasks
WHERE is_raw = true
  AND status = 'failed'
  AND failed_reason ~ 'duplicate #[0-9]+(:|\s)';

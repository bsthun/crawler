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
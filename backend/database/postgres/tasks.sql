-- name: TaskCreateForUserId :one
INSERT INTO tasks (user_id, category_id, type, url)
VALUES ($1, $2, $3, $4)
RETURNING *;
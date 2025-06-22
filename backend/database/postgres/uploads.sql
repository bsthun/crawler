-- name: UploadGetByIdAndUserId :one
SELECT *
FROM uploads
WHERE id = $1 AND user_id = $2;

-- name: UploadListByUserId :many
SELECT *
FROM uploads
WHERE user_id = $1
ORDER BY created_at DESC;
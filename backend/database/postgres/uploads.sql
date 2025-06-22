-- name: UploadGetByIdAndUserId :one
SELECT *
FROM uploads
WHERE id = $1 AND user_id = $2;
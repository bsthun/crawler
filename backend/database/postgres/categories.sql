-- name: CategoryGetByName :one
SELECT *
FROM categories
WHERE name = $1;
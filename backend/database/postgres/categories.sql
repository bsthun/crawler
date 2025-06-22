-- name: CategoryGetByName :one
SELECT *
FROM categories
WHERE name = $1;

-- name: CategoryList :many
SELECT *
FROM categories
ORDER BY name;
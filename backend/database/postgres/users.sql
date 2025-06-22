-- name: UserGetById :one
SELECT *
FROM users
WHERE id = $1;

-- name: UserGetByOid :one
SELECT *
FROM users
WHERE oid = $1;

-- name: UserCreate :one
INSERT INTO users (oid, firstname, lastname, email, photo_url)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

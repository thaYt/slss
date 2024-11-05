-- name: GetUser :one
SELECT *
FROM users
WHERE id = ? LIMIT 1;

-- name: GetUserByToken :one
SELECT *
FROM users
WHERE token = ? LIMIT 1;

-- name: GetUserByUsername :one
SELECT *
FROM users
WHERE username = ? LIMIT 1;

-- name: ListUsers :many
SELECT *
FROM users
ORDER BY username;

-- name: CreateUser :one
INSERT INTO users
    (
    username,
    password,
    token
    )
VALUES
    (
        ?, ?, ?
    )
RETURNING *;

-- name: SetUserPass :exec
UPDATE users
set username = ?,
password = ?
WHERE id = ?;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = ?;

-- name: GetFile :one
SELECT *
FROM files
WHERE id = ? LIMIT 1;

-- name: ListFiles :many
SELECT *
FROM files
ORDER BY path;

-- name: CreateFile :one
INSERT INTO files
    (
    alias,
    path,
    user_id,
    filetype,
    filesize
    )
VALUES
    (
        ?, ?, ?, ?, ?
    )
RETURNING *;

-- name: DeleteFile :exec
DELETE FROM files
WHERE id = ?;

-- name: GetFileByAlias :one
SELECT *
FROM files
WHERE alias = ? LIMIT 1;

-- name: GetFileByPath :one
SELECT *
FROM files
WHERE path = ? LIMIT 1;

-- name: GetFileByUser :many
SELECT *
FROM files
WHERE user_id = ?;

-- name: SetFileAlias :exec
UPDATE files
set alias = ?
WHERE id = ?;

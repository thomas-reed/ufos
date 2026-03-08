-- name: CreateObject :exec
INSERT INTO objects (id, parent_hash, is_folder, metadata, size_bytes, created_at)
VALUES (?, ?, ?, ?, ?, ?);

-- name: GetObject :one
SELECT * FROM objects WHERE id = ?;

-- name: ListObjectsByParent :many
SELECT * FROM objects WHERE parent_hash = ?;

-- name: DeleteObject :exec
DELETE FROM objects WHERE id = ?;
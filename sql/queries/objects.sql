-- name: CreateObject :one
INSERT INTO objects (
  id,
  prefix_hash,
  size_bytes,
  upload_status,
  metadata,
  created_at,
  updated_at
)
VALUES (
  ?,
  ?,
  ?,
  ?,
  ?,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
)
RETURNING id, created_at;

-- name: UpdateObject :one
UPDATE objects
SET
  prefix_hash = COALESCE(sqlc.narg('prefix_hash'), prefix_hash),
  size_bytes = COALESCE(sqlc.narg('size_bytes'), size_bytes),
  metadata = COALESCE(sqlc.narg('metadata'), metadata),
  upload_status = COALESCE(sqlc.narg('upload_status'), upload_status),
  updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING id, updated_at;

-- name: UpdateStatus :one
UPDATE objects
SET
  upload_status = ?,
  updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING id, upload_status;

-- name: GetObject :one
SELECT * FROM objects WHERE id = ?;

-- name: GetObjectsByParent :many
SELECT * FROM objects WHERE prefix_hash = ?;

-- name: DeleteObject :one
DELETE FROM objects WHERE id = ?
RETURNING id, size_bytes;

-- name: GetObjectsByTag :many
SELECT objects.*
FROM objects
INNER JOIN object_tags ON object_tags.object_id = objects.id
WHERE object_tags.tag_hash = ?;
-- name: CreateObject :exec
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
);

-- name: UpdateObject :exec
UPDATE objects
SET
  prefix_hash = ?,
  size_bytes = ?,
  upload_status = ?,
  metadata = ?,
  updated_at = CURRENT_TIMESTAMP
WHERE id = ?;

-- name: GetObject :one
SELECT * FROM objects WHERE id = ?;

-- name: ListObjectsByParent :many
SELECT * FROM objects WHERE prefix_hash = ?;

-- name: DeleteObject :exec
DELETE FROM objects WHERE id = ?;

-- name: GetObjectsByTag :many
SELECT objects.*
FROM objects
INNER JOIN object_tags ON object_tags.object_id = objects.id
WHERE object_tags.tag_hash = ?;
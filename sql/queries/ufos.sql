-- name: CreateUFO :one
INSERT INTO ufos (
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

-- name: UpdateUFO :one
UPDATE ufos
SET
  prefix_hash = COALESCE(sqlc.narg('prefix_hash'), prefix_hash),
  size_bytes = COALESCE(sqlc.narg('size_bytes'), size_bytes),
  metadata = COALESCE(sqlc.narg('metadata'), metadata),
  upload_status = COALESCE(sqlc.narg('upload_status'), upload_status),
  updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING id, updated_at;

-- name: UpdateStatus :one
UPDATE ufos
SET
  upload_status = ?,
  updated_at = CURRENT_TIMESTAMP
WHERE id = ?
RETURNING id, upload_status;

-- name: GetUFO :one
SELECT * FROM ufos WHERE id = ?;

-- name: GetUFOsByParent :many
SELECT * FROM ufos WHERE prefix_hash = ?;

-- name: DeleteUFO :one
DELETE FROM ufos WHERE id = ?
RETURNING id, size_bytes;

-- name: GetUFOsByTag :many
SELECT ufos.*
FROM ufos
INNER JOIN ufo_tags ON ufo_tags.ufo_id = ufos.id
WHERE ufo_tags.tag_hash = ?;
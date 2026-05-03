-- name: CreateUFO :one
INSERT INTO ufos (
  id,
  name_hash,
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
  ?,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
)
RETURNING id, created_at;

-- name: CreateUFOFolder :one
INSERT INTO ufos (
  id,
  name_hash,
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
  -1,
  "active",
  ?,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
)
ON CONFLICT(name_hash, prefix_hash) DO UPDATE SET
    updated_at = EXCLUDED.updated_at
RETURNING id, created_at, updated_at;

-- name: UpdateUFO :one
UPDATE ufos
SET
  name_hash = COALESCE(sqlc.narg('name_hash'), name_hash),
  prefix_hash = COALESCE(sqlc.narg('prefix_hash'), prefix_hash),
  size_bytes = COALESCE(sqlc.narg('size_bytes'), size_bytes),
  metadata = COALESCE(sqlc.narg('metadata'), metadata),
  upload_status = COALESCE(sqlc.narg('upload_status'), upload_status),
  updated_at = CURRENT_TIMESTAMP
WHERE id = sqlc.arg('id')
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

-- name: DeleteUFO :one
DELETE FROM ufos WHERE id = ?
RETURNING id, size_bytes;

-- name: GetUFOsByParent :many
SELECT * FROM ufos WHERE prefix_hash = ?;

-- name: GetUFOByNameAndParent :one
SELECT * FROM ufos WHERE name_hash = ? AND prefix_hash = ?;

-- name: GetUFOsByTags :many
SELECT ufos.*
FROM ufos
INNER JOIN ufo_tags ON ufo_tags.ufo_id = ufos.id
WHERE ufo_tags.tag_hash IN (sqlc.slice('tags'))
GROUP BY ufos.id
HAVING COUNT(DISTINCT ufo_tags.tag_hash) = CAST(sqlc.arg('tag_count') AS INTEGER);

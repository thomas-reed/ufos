-- name: AddToOrbit :one
INSERT INTO orbit (
  persona_id,
  signing_key,
  exchange_key,
  metadata,
  created_at,
  updated_at
)
VALUES (
  ?,
  ?,
  ?,
  ?,
  CURRENT_TIMESTAMP,
  CURRENT_TIMESTAMP
)
RETURNING persona_id, created_at;

-- name: UpdateOrbitMetadata :one
UPDATE orbit
SET
  metadata = ?,
  updated_at = CURRENT_TIMESTAMP
WHERE persona_id = ?
RETURNING persona_id, updated_at;

-- name: GetOrbitList :many
SELECT * FROM orbit;

-- name: GetFromOrbit :one
SELECT * FROM orbit WHERE persona_id = ?;

-- name: DeleteFromOrbit :one
DELETE FROM orbit WHERE persona_id = ?
RETURNING persona_id;
-- name: AddUFOAccess :one
INSERT OR IGNORE INTO ufo_access (ufo_id, persona_id, wrapped_key)
VALUES (
  ?,
  ?,
  ?
)
RETURNING ufo_id;

-- name: DeleteUFOAccess :exec
DELETE FROM ufo_access WHERE ufo_id = ?;

-- name: GetUsersForUFO :many
SELECT * FROM ufo_access WHERE ufo_id = ?;

-- name: GetKeybyUFOIDAndPersonaID :one
SELECT wrapped_key FROM ufo_access
WHERE ufo_id = ? AND persona_id = ?;

-- name: GetUFOAccessForUser :one
SELECT COUNT(*) FROM ufo_access
WHERE ufo_id = ? AND persona_id = ?;
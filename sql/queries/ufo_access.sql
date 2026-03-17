-- name: AddUFOAccess :one
INSERT OR IGNORE INTO ufo_access (ufo_id, persona_id)
VALUES (
  ?,
  ?
)
RETURNING ufo_id;

-- name: DeleteUFOAccess :exec
DELETE FROM ufo_access WHERE ufo_id = ?;

-- name: GetUsersForUFO :many
SELECT * FROM ufo_access WHERE ufo_id = ?;

-- name: GetUFOAccessForUser :one
SELECT COUNT(*) FROM ufo_access
WHERE ufo_id = ? AND persona_id = ?;
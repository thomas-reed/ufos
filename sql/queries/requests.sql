-- name: NewRequest :exec
INSERT INTO requests (id)
VALUES (?);

-- name: GetRequestByID :one
SELECT * FROM requests WHERE id = ?;

-- name: DeleteStaleRequests :exec
DELETE FROM requests WHERE created_at < DATETIME('now', '-5 minutes');
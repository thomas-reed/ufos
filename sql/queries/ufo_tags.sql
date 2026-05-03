-- name: AddUFOTag :exec
INSERT OR IGNORE INTO ufo_tags (ufo_id, tag_hash)
VALUES (?, ?);

-- name: DeleteUFOTags :exec
DELETE FROM ufo_tags WHERE ufo_id = ?;

-- name: GetTagsForUFO :many
SELECT * FROM ufo_tags WHERE ufo_id = ?;
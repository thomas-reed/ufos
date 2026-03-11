-- name: AddObjectTag :one
INSERT INTO object_tags (object_id, tag_hash)
VALUES (
  ?,
  ?
)
RETURNING object_id;

-- name: DeleteObjectTags :exec
DELETE FROM object_tags WHERE object_id = ?;

-- name: GetTagsForObject :many
SELECT * FROM object_tags WHERE object_id = ?;
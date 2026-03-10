-- +goose Up

-- The objects table tracks the encrypted files stored on the server.
CREATE TABLE objects (
    id TEXT PRIMARY KEY,            -- Unique UUID for the file
    prefix_hash TEXT NOT NULL,      -- Hashed prefix for folder-like navigation
    size_bytes INTEGER NOT NULL,    -- size of file on disk - "folder" will be 0
    upload_status TEXT NOT NULL,    -- pending, uploading, active, failed
    metadata BLOB,                  -- Encrypted JSON (filename, type, etc.)
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
) WITHOUT ROWID;

-- The object_tags table allows for searching by many hashed keywords.
CREATE TABLE object_tags (
    object_id TEXT NOT NULL,
    tag_hash TEXT NOT NULL,
    PRIMARY KEY (object_id, tag_hash),
    FOREIGN KEY (object_id) REFERENCES objects(id) ON DELETE CASCADE
) WITHOUT ROWID;

-- Index for searching specifically by a hashed tag.
CREATE INDEX idx_tag_hash ON object_tags(tag_hash);

-- The requests table provides replay protection for all API calls.
CREATE TABLE requests (
    id TEXT PRIMARY KEY,               -- Unique Request UUID
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) WITHOUT ROWID;

-- +goose Down
DROP TABLE requests;
DROP TABLE object_tags;
DROP TABLE objects;

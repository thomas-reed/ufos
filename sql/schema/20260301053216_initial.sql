-- +goose Up

-- Tracks the encrypted files stored on the server.
CREATE TABLE ufos (
    id TEXT PRIMARY KEY,            -- Unique UUID for the file
    name_hash TEXT NOT NULL,        -- Hashed name of UFO for determining uniqueness inside a "folder"
    prefix_hash TEXT NOT NULL,      -- Hashed prefix for folder-like navigation
    size_bytes INTEGER NOT NULL,    -- size of file on disk - "folder" will be -1
    upload_status TEXT NOT NULL,    -- pending, uploading, active, failed
    metadata BLOB NOT NULL,         -- Encrypted JSON (filename, type, etc.)
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    UNIQUE(prefix_hash, name_hash)  -- Ensures 2 "files" of the with the same name can't be in the same "folder" 
) WITHOUT ROWID;

-- Allows for searching by many hashed keywords.
CREATE TABLE ufo_tags (
    ufo_id TEXT NOT NULL,
    tag_hash TEXT NOT NULL,
    PRIMARY KEY (ufo_id, tag_hash),
    FOREIGN KEY (ufo_id) REFERENCES ufos(id) ON DELETE CASCADE
) WITHOUT ROWID;

-- Index for searching specifically by a hashed tag.
CREATE INDEX idx_tag_hash ON ufo_tags(tag_hash);

-- Stores the public keys of trusted acquaintances.
CREATE TABLE orbit (
    persona_id TEXT PRIMARY KEY, -- The ID of the friend
    signing_key BLOB NOT NULL,    -- Their ed25519 public key for signature verification
    exchange_key BLOB NOT NULL,  -- Their x25519 public key for encryption
    metadata BLOB,               -- Optional encrypted info about user
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL
) WITHOUT ROWID;

-- Provides server-side authorization for a given ufo
CREATE TABLE ufo_access (
    ufo_id TEXT NOT NULL,
    persona_id TEXT NOT NULL,
    wrapped_key BLOB NOT NULL,
    PRIMARY KEY (ufo_id, persona_id),
    FOREIGN KEY (ufo_id) REFERENCES ufos(id) ON DELETE CASCADE,
    FOREIGN KEY (persona_id) REFERENCES orbit(persona_id) ON DELETE CASCADE
) WITHOUT ROWID;

-- Provides replay protection for all API calls.
CREATE TABLE requests (
    id TEXT PRIMARY KEY,               -- Unique Request UUID
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
) WITHOUT ROWID;

-- +goose Down
DROP TABLE requests;
DROP TABLE ufo_access;
DROP TABLE ufo_tags;
DROP TABLE ufos;

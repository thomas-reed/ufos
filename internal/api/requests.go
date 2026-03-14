package api

import "crypto/ed25519"

// for use with CreateUFO, UpdateUFO
type UFOMetadataRequest struct {
	PrefixHash string `json:"prefix_hash"`
	SizeBytes  int64  `json:"size_bytes"`
	Metadata   []byte `json:"metadata"`
}

// for use with CreatePersona
type NewPersonaRequest struct {
	ID        string            `json:"id"`
	PublicKey ed25519.PublicKey `json:"public_key"`
}

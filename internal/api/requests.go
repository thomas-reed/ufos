package api

import "crypto/ed25519"

const (
	RouteHealth   = "/healthz"
	RouteInit     = "/api/init"
	RouteRegister = "/api/personas"
	RouteUFOs     = "/api/ufos"
	RouteOrbit    = "/api/orbit"
	RouteSearch   = "/api/tags"

	HeaderRegistration = "X-UFO-Registration"
	HeaderTimestamp    = "X-UFO-Timestamp"
	HeaderPersona      = "X-UFO-Persona"
	HeaderHost         = "X-UFO-Host"
	HeaderSignature    = "X-UFO-Signature"
)

// for use with CreateUFO, UpdateUFO
type UFOMetadataRequest struct {
	PrefixHash string            `json:"prefix_hash"`
	SizeBytes  int64             `json:"size_bytes"`
	Metadata   []byte            `json:"metadata"`
	AccessList map[string][]byte `json:"access_list"` // map[persona_id]wrapped_key
	TagHashes  []string          `json:"tag_hashes"`
}

// for use with CreatePersona
type NewPersonaRequest struct {
	ID        string            `json:"id"`
	PublicKey ed25519.PublicKey `json:"public_key"`
}

type OrbitMetadataRequest struct {
	PersonaID string            `json:"persona_id"`
	PublicKey ed25519.PublicKey `json:"public_key"`
	Metadata  []byte            `json:"metadata"`
}

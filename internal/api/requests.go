package api

const (
	RouteHealth   = "/healthz"
	RouteInit     = "/api/init"
	RoutePersonas = "/api/personas"
	RouteUFOs     = "/api/ufos"
	RouteOrbit    = "/api/orbit"

	HeaderRegistration = "X-UFO-Registration"
	HeaderTimestamp    = "X-UFO-Timestamp"
	HeaderPersona      = "X-UFO-Persona"
	HeaderHost         = "X-UFO-Host"
	HeaderSignature    = "X-UFO-Signature"
)

// for use with CreateUFO, UpdateUFO
type UFOMetadataRequest struct {
	NameHash   *string           `json:"name_hash"`
	PrefixHash *string           `json:"prefix_hash"`
	SizeBytes  *int64            `json:"size_bytes"`
	Metadata   []byte            `json:"metadata"`
	AccessList map[string][]byte `json:"access_list"` // map[persona_id]wrapped_key
	TagHashes  []string          `json:"tag_hashes"`
}

// for use with CreatePersona
type NewPersonaRequest struct {
	ID          string `json:"id"`
	SigningKey  []byte `json:"signing_key"`  // Ed25519 public key
	ExchangeKey []byte `json:"exchange_key"` // X25519 public key
}

type OrbitMetadataRequest struct {
	PersonaID   string `json:"persona_id"`   // Persona ID of your contact
	SigningKey  []byte `json:"signing_key"`  // their Ed25519 public key
	ExchangeKey []byte `json:"exchange_key"` // their X25519 public key
	Metadata    []byte `json:"metadata"`     // encrypted blob of metadata
}

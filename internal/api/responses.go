package api

import (
	"net/http"
	"time"

	"github.com/thomas-reed/ufos/internal/objects"
)

const (
	HeaderMetadata   = "X-UFO-Metadata"
	HeaderWrappedKey = "X-UFO-Wrapped-Key"
)

type CreateUFOResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateUFOResponse struct {
	ID        string    `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UploadUFOResponse struct {
	ID     string            `json:"id"`
	Status objects.UFOStatus `json:"status"`
}

// List, Search return a []UFOItem
type UFOItem struct {
	ID         string            `json:"id"`
	PrefixHash string            `json:"prefix_hash"`
	SizeBytes  int64             `json:"size_bytes"`
	Status     objects.UFOStatus `json:"status"`
	Metadata   []byte            `json:"metadata"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

type InitPersonaResponse struct {
	RegistrationToken string `json:"registration_token"`
}

type CreatePersonaResponse struct {
	ID string `json:"ID"`
}

type OrbitItem struct {
	PersonaID   string    `json:"persona_id"`
	SigningKey  []byte    `json:"signing_key"`
	ExchangeKey []byte    `json:"exchange_key"`
	Metadata    []byte    `json:"metadata"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type AddToOrbitResponse struct {
	PersonaID string    `json:"persona_id"`
	CreatedAt time.Time `json:"created_at"`
}

type HeaderDecoder interface {
	DecodeHeader(http.Header) error
}

type UFOMetadataFromHeader struct {
	MetadataBlob []byte `json:"metadataBlob"`
}

func (u *UFOMetadataFromHeader) DecodeHeader(h http.Header) error {
	u.MetadataBlob = []byte(h.Get(HeaderMetadata))
	return nil
}

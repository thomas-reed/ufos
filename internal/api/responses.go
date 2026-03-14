package api

import (
	"time"

	"github.com/thomas-reed/ufos/internal/objects"
)

type CreateUFOResponse struct {
	ID        string    `json:"id"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateUFOResponse struct {
	ID        string    `json:"id"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UploadObjectResponse struct {
	ID     string               `json:"id"`
	Status objects.ObjectStatus `json:"status"`
}

// List, Search return a []UFOItem
type UFOItem struct {
	ID         string               `json:"id"`
	PrefixHash string               `json:"prefix_hash"`
	SizeBytes  int64                `json:"size_bytes"`
	Status     objects.ObjectStatus `json:"status"`
	Metadata   []byte               `json:"metadata"`
	CreatedAt  time.Time            `json:"created_at"`
	UpdatedAt  time.Time            `json:"updated_at"`
}

type InitPersonaResponse struct {
	RegistrationToken string `json:"registration_token"`
}

type CreatePersonaResponse struct {
	ID string `json:"ID"`
}

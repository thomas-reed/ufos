package api

import (
	"time"

	"github.com/thomas-reed/ufos/internal/objects"
)

type CreateUFOResponse struct {
	ID				string 		`json:"id"`
	CreatedAt	time.Time	`json:"created_at"`
}

type UpdateUFOResponse struct {
	ID				string 		`json:"id"`
	UpdatedAt	time.Time	`json:"updated_at"`
}

type UploadObjectResponse struct {
	ID			string								`json:"id"`
	Status	objects.ObjectStatus	`json:"status"`
}
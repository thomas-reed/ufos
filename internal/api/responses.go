package api

import "time"

type CreateUFOResponse struct {
	ID				string 		`json:"id"`
	CreatedAt	time.Time	`json:"created_at"`
}

type UpdateUFOResponse struct {
	ID				string 		`json:"id"`
	UpdatedAt	time.Time	`json:"updated_at"`
}
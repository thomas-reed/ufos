package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/database"
	"github.com/thomas-reed/ufos/internal/objects"
)

func (s *Server) HandleCreateUFO(w http.ResponseWriter, r *http.Request) {
	p, ok := r.Context().Value(personaKey).(*Persona)
	if !ok {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"couldn't get persona from context",
			nil,
		)
		return
	}
	ufoID := uuid.New().String()

	var req api.UFOMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid json body", err)
		return
	}

	status := objects.StatusPending
	if req.SizeBytes <= 0 {
		status = objects.StatusActive
	}

	params := database.CreateUFOParams{
		ID:           ufoID,
		PrefixHash:   req.PrefixHash,
		SizeBytes:    req.SizeBytes,
		UploadStatus: string(status),
		Metadata:     req.Metadata,
	}
	// Set up db transaction
	tx, err := p.DBConn.BeginTx(r.Context(), nil)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"db transaction setup failed",
			err,
		)
		return
	}
	defer tx.Rollback()
	qtx := p.db.WithTx(tx)

	// Create with the transaction
	res, err := qtx.CreateUFO(r.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create ufo", err)
		return
	}

	// Add new tag hash references to ufo
	for _, tag := range req.Tags {
		_, err := qtx.AddUFOTag(r.Context(), database.AddUFOTagParams{
			UfoID:   res.ID,
			TagHash: tag,
		})
		if err != nil {
			respondWithError(
				w,
				http.StatusInternalServerError,
				"failed to create ufo tag",
				err,
			)
			return
		}
	}

	// Add any personas to the Access List
	for _, recipient := range req.AccessList {
		if _, err := qtx.AddUFOAccess(r.Context(), database.AddUFOAccessParams{
			UfoID:     res.ID,
			PersonaID: recipient,
		}); err != nil {
			respondWithError(
				w,
				http.StatusInternalServerError,
				"failed to add permission",
				err,
			)
			return
		}
	}

	// Submit transaction
	err = tx.Commit()
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"db transaction failed",
			err,
		)
		return
	}

	jsonResponse(w, http.StatusCreated, api.CreateUFOResponse{
		ID:        res.ID,
		CreatedAt: res.CreatedAt,
	},
	)
}

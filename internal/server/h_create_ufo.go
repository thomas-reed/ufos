package server

import (
	"encoding/json"
	"github.com/google/uuid"
	"net/http"

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

	params := database.CreateObjectParams{
		ID: ufoID,
		PrefixHash: req.PrefixHash,
		SizeBytes: req.SizeBytes,
		UploadStatus: string(status),
		Metadata: req.Metadata,
	}

	res, err := p.db.CreateObject(r.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create object", err)
		return
	}
	jsonResponse(
		w,
		http.StatusCreated,
		api.CreateUFOResponse{
			ID: res.ID,
			CreatedAt: res.CreatedAt,
		},
	)
}
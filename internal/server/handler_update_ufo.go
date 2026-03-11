package server

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/database"
)

func (s *Server) HandleUpdateUFO(w http.ResponseWriter, r *http.Request) {
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
	ufoID := r.PathValue("id")

	var req api.UFOMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid json body", err)
		return
	}

	// Convert your request fields to Nullable types
	params := database.UpdateObjectParams{
		ID: ufoID,
		PrefixHash: sql.NullString{String: req.PrefixHash, Valid: req.PrefixHash != ""},
		SizeBytes:  sql.NullInt64{Int64: req.SizeBytes, Valid: req.SizeBytes != 0},
		Metadata:   req.Metadata,
		UploadStatus: sql.NullString{Valid: false},
	}

	updated, err := p.db.UpdateObject(r.Context(), params)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to update object", err)
		return
	}
	
	jsonResponse(w, http.StatusOK, updated)
}

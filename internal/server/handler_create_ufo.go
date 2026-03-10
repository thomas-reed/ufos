package server

import (
	"encoding/base64"
	"github.com/google/uuid"
	"net/http"
	"strconv"

	"github.com/thomas-reed/ufos/internal/database"
	"github.com/thomas-reed/ufos/internal/objects"
)
type CreateUFOResponse struct {
	ID string `json:"id"`
}

func (s *Server) HandleCreateUFO(w http.ResponseWriter, r *http.Request) {
	p, ok := r.Context().Value(personaKey).(*Persona)
	if !ok {
		respondWithError(w, 500, "internal server error", nil)
		return
	}
	ufoID := uuid.New().String()
	prefixHashStr := r.Header.Get("X-UFO-Prefix")
	size, _ := strconv.ParseInt(r.Header.Get("X-UFO-Size"), 10, 64)
	metadata, err := base64.StdEncoding.DecodeString(r.Header.Get("X-UFO-Metadata"))
	if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid metadata encoding", err)
			return
	}
	status := objects.StatusPending
	if size == 0 {
		status = objects.StatusActive
	}

	params := database.CreateObjectParams{
		ID: ufoID,
		PrefixHash: prefixHashStr,
		SizeBytes: size,
		UploadStatus: string(status),
		Metadata: metadata,
	}

	if err := p.db.CreateObject(r.Context(), params); err != nil {
		respondWithError(w, http.StatusInternalServerError, "failed to create object", err)
		return
	}
	jsonResponse(w, http.StatusCreated, CreateUFOResponse{ ID: ufoID })
}
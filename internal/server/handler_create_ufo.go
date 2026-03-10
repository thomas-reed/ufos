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
	if prefixHashStr == "" {
			respondWithError(w, http.StatusBadRequest, "missing object prefix", nil)
			return
	}
	
	sizeStr := r.Header.Get("X-UFO-Size")
	if sizeStr == "" {
			respondWithError(w, http.StatusBadRequest, "missing object size", nil)
			return
	}
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid size format", err)
			return
	}

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
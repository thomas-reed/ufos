package server

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/thomas-reed/ufos/internal/api"
)

func (s *Server) HandleGetUFOMetadata(w http.ResponseWriter, r *http.Request) {
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

	ufoID := r.PathValue("uuid")

	ufo, err := p.db.GetUFO(r.Context(), ufoID)
	if err != nil {
		respondWithError(
			w, http.StatusNotFound,
			"couldn't retrieve ufo from db",
			err,
		)
		return
	}

	// Construct headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", ufo.SizeBytes))
	w.Header().Set(api.HeaderMetadata, base64.StdEncoding.EncodeToString(ufo.Metadata))

	w.WriteHeader(http.StatusOK)
}

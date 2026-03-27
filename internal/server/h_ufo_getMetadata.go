package server

import (
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/database"
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
	if r.Header.Get(api.HeaderHost) == "" {
		// Owner Path
		w.Header().Set(api.HeaderMetadata, base64.StdEncoding.EncodeToString(ufo.Metadata))
	} else {
		// Guest Path: Fetch and send the Wrapped-Key header
		wrappedKey, err := p.db.GetKeybyUFOIDAndPersonaID(
			r.Context(),
			database.GetKeybyUFOIDAndPersonaIDParams{
				UfoID:     ufoID,
				PersonaID: p.ID,
			},
		)
		if err == nil {
			w.Header().Set(api.HeaderWrappedKey, base64.StdEncoding.EncodeToString(wrappedKey))
		}
	}

	w.WriteHeader(http.StatusOK)
}

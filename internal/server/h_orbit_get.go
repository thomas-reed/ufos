package server

import (
	"net/http"

	"github.com/thomas-reed/ufos/internal/api"
)

func (s *Server) HandleGetFromOrbit(w http.ResponseWriter, r *http.Request) {
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

	personaID := r.PathValue("id")

	personaData, err := p.db.GetFromOrbit(r.Context(), personaID)
	if err != nil {
		respondWithError(
			w, http.StatusNotFound,
			"Error retrieving user from orbit",
			err,
		)
		return
	}

	jsonResponse(
		w,
		http.StatusOK,
		api.OrbitItem{
			PersonaID:   personaData.PersonaID,
			SigningKey:  personaData.SigningKey,
			ExchangeKey: personaData.ExchangeKey,
			Metadata:    personaData.Metadata,
			CreatedAt:   personaData.CreatedAt,
			UpdatedAt:   personaData.UpdatedAt,
		},
	)
}

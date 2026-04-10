package server

import (
	"net/http"

	"github.com/thomas-reed/ufos/internal/api"
)

func (s *Server) HandleGetPersonaKeys(w http.ResponseWriter, r *http.Request) {
	personaID := r.PathValue("id")

	persona, ok := s.GetPersona(personaID)
	if !ok {
		respondWithError(w, http.StatusNotFound, "Unknown persona ID", nil)
	}

	jsonResponse(w, http.StatusOK, api.PersonaKeysResponse{
		PersonaID: persona.ID,
		SigningKey: persona.SigningKey,
		ExchangeKey: persona.ExchangeKey,
	})
}

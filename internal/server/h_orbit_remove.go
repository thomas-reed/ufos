package server

import (
	"net/http"
)

func (s *Server) HandleRemoveFromOrbit(w http.ResponseWriter, r *http.Request) {
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

	_, err := p.db.DeleteFromOrbit(r.Context(), personaID)
	if err != nil {
		respondWithError(
			w, http.StatusNotFound,
			"Error removing user from orbit",
			err,
		)
		return
	}

	jsonResponse(
		w,
		http.StatusNoContent,
		nil,
	)
}

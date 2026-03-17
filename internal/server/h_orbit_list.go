package server

import (
	"net/http"

	"github.com/thomas-reed/ufos/internal/api"
)

func (s *Server) HandleOrbitList(w http.ResponseWriter, r *http.Request) {
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
	list, err := p.db.GetOrbitList(r.Context())
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"Error getting orbit list",
			err,
		)
		return
	}

	orbitList := make([]api.OrbitItem, 0, len(list))
	for i := range list {
		orbitList = append(orbitList, api.OrbitItem{
			PersonaID: list[i].PersonaID,
			PublicKey: list[i].PublicKey,
			Metadata:  list[i].Metadata,
			CreatedAt: list[i].CreatedAt,
			UpdatedAt: list[i].UpdatedAt,
		})
	}

	jsonResponse(w, http.StatusOK, orbitList)
}

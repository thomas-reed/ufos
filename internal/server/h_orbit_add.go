package server

import (
	"encoding/json"
	"net/http"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/database"
)

func (s *Server) HandleAddToOrbit(w http.ResponseWriter, r *http.Request) {
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

	var req api.OrbitMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid json body", err)
		return
	}

	res, err := p.db.AddToOrbit(
		r.Context(),
		database.AddToOrbitParams{
			PersonaID:   req.PersonaID,
			SigningKey:  req.SigningKey,
			ExchangeKey: req.ExchangeKey,
			Metadata:    req.Metadata,
		},
	)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"failed to add user to orbit",
			err,
		)
		return
	}
	jsonResponse(w, http.StatusCreated, api.AddToOrbitResponse{
		PersonaID: res.PersonaID,
		CreatedAt: res.CreatedAt,
	},
	)
}

package server

import (
	"encoding/base64"
	"net/http"
	"time"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/crypto"
)

func (s *Server) HandleInitPersona(w http.ResponseWriter, r *http.Request) {
	key, err := crypto.GenerateKey()
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"error generating token",
			err,
		)
		return
	}
	s.mu.Lock()
	s.registrationToken = base64.StdEncoding.EncodeToString(key)
	s.tokenCreated = time.Now().UTC()
	s.mu.Unlock()

	jsonResponse(w, http.StatusCreated, api.InitPersonaResponse{
		RegistrationToken: s.registrationToken,
	})
}

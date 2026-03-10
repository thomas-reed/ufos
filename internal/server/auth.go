package server

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/thomas-reed/ufos/internal/crypto"
)

type contextKey string

const personaKey contextKey = "persona"

func (s *Server) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		personaID := r.Header.Get("X-UFO-Persona")
		timestampStr := r.Header.Get("X-UFO-Timestamp")
		sizeStr := r.Header.Get("X-UFO-Size")
		prefixHashStr := r.Header.Get("X-UFO-Prefix")
		metadataStr := r.Header.Get("X-UFO-Metadata")
		sigBase64 := r.Header.Get("X-UFO-Signature")

		timestampInt, err := strconv.ParseInt(timestampStr, 10, 64)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "invalid timestamp", nil)
			return
		}
		requestTime := time.Unix(timestampInt, 0).UTC()
		now := time.Now().UTC()
		delta := now.Sub(requestTime)
		if delta < 0 {
			delta = -delta
		}

		if delta > 5*time.Minute {
			respondWithError(w, http.StatusUnauthorized, "timestamp out of range", nil)
			return
		}

		persona, ok := s.GetPersona(personaID)
		if !ok {
			respondWithError(w, http.StatusUnauthorized, "unauthorized", nil)
			return
		}

		payload := fmt.Sprintf(
			"%s|%s|%s|%s|%s|%s|%s",
			r.Method,
			r.URL.Path,
			personaID,
			timestampStr,
			sizeStr,
			prefixHashStr,
			metadataStr,
		)

		sig, _ := base64.StdEncoding.DecodeString(sigBase64)

		if !crypto.VerifyRequest(persona.PublicKey, []byte(payload), sig) {
			respondWithError(w, http.StatusUnauthorized, "invalid signature", nil)
			return
		}

		if _, err := persona.db.GetRequestByID(r.Context(), sigBase64); err == sql.ErrNoRows {
			persona.db.NewRequest(r.Context(), sigBase64)
		} else if err != nil {
			respondWithError(w, http.StatusInternalServerError, "database error", err)
			return
		} else {
			respondWithError(w, http.StatusUnauthorized, "duplicate signature", nil)
			return
		}

		ctx := context.WithValue(r.Context(), personaKey, persona)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

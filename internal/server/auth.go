package server

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/crypto"
	"github.com/thomas-reed/ufos/internal/database"
)

type contextKey string

const personaKey contextKey = "persona"

// Authenticate provides the standard gate for signed JSON requests
func (s *Server) Authenticate(next http.Handler) http.Handler {
	return s.auth(next, true)
}

// AuthenticateStream provides the gate for the binary ufos
func (s *Server) AuthenticateStream(next http.Handler) http.Handler {
	return s.auth(next, false)
}

func (s *Server) auth(next http.Handler, requiresBodyHash bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the Timestamp
		timestampStr := r.Header.Get(api.HeaderTimestamp)
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

		// Check the Persona
		personaID := r.Header.Get(api.HeaderPersona)
		persona, ok := s.GetPersona(personaID)
		if !ok {
			// Ensure it must be a GET or HEAD for /api/ufos/{uuid}
			ufoID := r.PathValue("uuid")
			isAllowedMethod := r.Method == http.MethodGet || r.Method == http.MethodHead
			isCorrectPath := strings.Contains(r.URL.Path, "/api/ufos/")
			if !isAllowedMethod || !isCorrectPath || ufoID == "" {
				respondWithError(w, http.StatusUnauthorized, "unauthorized", nil)
				return
			}
			// Check if it's an authorized guest
			hostID := r.Header.Get(api.HeaderHost)
			if hostID == "" {
				respondWithError(w, http.StatusUnauthorized, "unauthorized", nil)
				return
			}
			host, found := s.GetPersona(hostID)
			if !found {
				respondWithError(w, http.StatusUnauthorized, "unauthorized", nil)
				return
			}
			// Ensure the guest has access for that particular UFO
			count, err := host.db.GetUFOAccessForUser(
				r.Context(),
				database.GetUFOAccessForUserParams{
					UfoID:     ufoID,
					PersonaID: personaID,
				},
			)
			if err != nil {
				respondWithError(
					w,
					http.StatusInternalServerError,
					"Error getting authorization from db",
					err,
				)
				return
			}
			if count < 1 {
				respondWithError(w, http.StatusUnauthorized, "unauthorized", nil)
				return
			}
			// Guest is now authorized, merge guest and host info into persona object for hash check
			personaData, err := host.db.GetOrbit(r.Context(), personaID)
			if err != nil {
				respondWithError(
					w,
					http.StatusInternalServerError,
					"Error getting user data from db",
					err,
				)
				return
			}
			persona = &Persona{
				ID:          personaData.PersonaID,
				SigningKey:  personaData.SigningKey,
				ExchangeKey: personaData.ExchangeKey,
				RootFS:      host.RootFS,
				DBPath:      host.DBPath,
				DBConn:      host.DBConn,
				db:          host.db,
			}
		}

		// Check the Hashed body (if required)
		sigBase64 := r.Header.Get(api.HeaderSignature)
		bodyHash := ""
		if requiresBodyHash && r.Body != nil {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				respondWithError(w, http.StatusInternalServerError, "could not read body", err)
				return
			}
			r.Body.Close()
			bodyHash = crypto.HashAndBase64(body)
			// Re-stuff the body so HandleCreateUFO can decode the JSON
			r.Body = io.NopCloser(bytes.NewBuffer(body))
		}

		payload := fmt.Sprintf(
			"%s|%s|%s|%s|%s",
			r.Method,
			r.URL.Path,
			personaID,
			timestampStr,
			bodyHash,
		)

		sig, _ := base64.StdEncoding.DecodeString(sigBase64)

		if !crypto.VerifyRequest(persona.SigningKey, []byte(payload), sig) {
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

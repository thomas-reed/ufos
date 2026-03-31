package server

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/mattn/go-sqlite3"
	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/database"
)

func (s *Server) HandleUpdateUFO(w http.ResponseWriter, r *http.Request) {
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

	var req api.UFOMetadataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid json body", err)
		return
	}

	ufoID := r.PathValue("uuid")
	params := database.UpdateUFOParams{
		ID:           ufoID,
		UploadStatus: sql.NullString{Valid: false},
		Metadata:     req.Metadata,
	}
	if req.NameHash != nil {
		params.NameHash = sql.NullString{
			String: *req.NameHash,
			Valid:  true,
		}
	}
	if req.PrefixHash != nil {
		params.PrefixHash = sql.NullString{
			String: *req.PrefixHash,
			Valid:  true,
		}
	}
	if req.SizeBytes != nil {
		params.SizeBytes = sql.NullInt64{
			Int64: *req.SizeBytes,
			Valid: true,
		}
	}

	// Set up db transaction
	tx, err := p.DBConn.BeginTx(r.Context(), nil)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"db transaction setup failed",
			err,
		)
		return
	}
	defer tx.Rollback()
	qtx := p.db.WithTx(tx)

	// Update with the transaction
	updated, err := qtx.UpdateUFO(r.Context(), params)
	if err != nil {
		// Check if error is a unique constraint violation
		if sqliteErr, ok := err.(sqlite3.Error); ok {
			if sqliteErr.Code == sqlite3.ErrConstraint {
				respondWithError(
					w,
					http.StatusConflict,
					"UFO with that name and prefix already exists",
					nil,
				)
				return
			}
		}
		// Otherwise fallback for all other errors
		respondWithError(w, http.StatusInternalServerError, "failed to update ufo", err)
		return
	}

	// Remove current tags and add new tag hash references to ufo
	if err = qtx.DeleteUFOTags(r.Context(), updated.ID); err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"failed to create ufo tag",
			err,
		)
		return
	}
	for _, tag := range req.TagHashes {
		_, err := qtx.AddUFOTag(r.Context(), database.AddUFOTagParams{
			UfoID:   updated.ID,
			TagHash: tag,
		})
		if err != nil {
			respondWithError(
				w,
				http.StatusInternalServerError,
				"failed to create ufo tag",
				err,
			)
			return
		}
	}

	// Remove current Access List and add given personas
	if err = qtx.DeleteUFOAccess(r.Context(), updated.ID); err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"failed to remove ufo access",
			err,
		)
		return
	}
	for recipient, wrappedKey := range req.AccessList {
		if _, err := qtx.AddUFOAccess(r.Context(), database.AddUFOAccessParams{
			UfoID:      updated.ID,
			PersonaID:  recipient,
			WrappedKey: wrappedKey,
		}); err != nil {
			respondWithError(
				w,
				http.StatusInternalServerError,
				"failed to add permission",
				err,
			)
			return
		}
	}

	// Submit transaction
	err = tx.Commit()
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"db transaction failed",
			err,
		)
		return
	}

	jsonResponse(w, http.StatusOK, api.UpdateUFOResponse{
		ID:        updated.ID,
		UpdatedAt: updated.UpdatedAt,
	})
}

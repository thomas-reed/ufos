package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/mattn/go-sqlite3"
	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/database"
	"github.com/thomas-reed/ufos/internal/objects"
)

func (s *Server) HandleCreateUFO(w http.ResponseWriter, r *http.Request) {
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

	ufoID := uuid.New().String()
	ufoRes := api.CreateUFOResponse{}
	var status int

	// Create with the transaction
	if *req.SizeBytes < 0 {
		// FOLDER BRANCH
		params := database.CreateUFOFolderParams{
			ID:         ufoID,
			NameHash:   *req.NameHash,
			PrefixHash: *req.PrefixHash,
			Metadata:   req.Metadata,
		}
		ufo, err := qtx.CreateUFOFolder(r.Context(), params)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Couldn't create folder", err)
			return
		}
		// Add new tag hash references to ufo
		if err = addTagRefs(r.Context(), qtx, ufo.ID, req.TagHashes); err != nil {
			respondWithError(
				w,
				http.StatusInternalServerError,
				"failed to create ufo tag",
				err,
			)
			return
		}
		status = http.StatusCreated		// assuming folder was created
		if ufo.CreatedAt != ufo.UpdatedAt {
			status = http.StatusOK			// folder already existed
		}
		ufoRes.ID = ufo.ID
		ufoRes.CreatedAt = ufo.CreatedAt
	} else {
		// FILE BRANCH
		params := database.CreateUFOParams{
			ID:           ufoID,
			NameHash:     *req.NameHash,
			PrefixHash:   *req.PrefixHash,
			SizeBytes:    *req.SizeBytes,
			UploadStatus: string(objects.StatusPending),
			Metadata:     req.Metadata,
		}
		ufo, err := qtx.CreateUFO(r.Context(), params)
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
			respondWithError(w, http.StatusInternalServerError, "failed to create ufo", err)
			return
		}
		// Add new tag hash references to ufo
		if err = addTagRefs(r.Context(), qtx, ufo.ID, req.TagHashes); err != nil {
			respondWithError(
				w,
				http.StatusInternalServerError,
				"failed to create ufo tag",
				err,
			)
			return
		}
		// Add any personas to the Access List
		for recipient, wrappedKey := range req.AccessList {
			if _, err := qtx.AddUFOAccess(r.Context(), database.AddUFOAccessParams{
				UfoID:      ufo.ID,
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
		status = http.StatusCreated
		ufoRes.ID = ufo.ID
		ufoRes.CreatedAt = ufo.CreatedAt
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

	jsonResponse(w, status, ufoRes)
}

func addTagRefs(
	ctx context.Context,
	qtx *database.Queries,
	ufoID string,
	tagHashes []string,
) error {
	for _, tag := range tagHashes {
		if _, err := qtx.AddUFOTag(ctx, database.AddUFOTagParams{
			UfoID:   ufoID,
			TagHash: tag,
		}); err != nil {
			return err
		}
	}
	return nil
}

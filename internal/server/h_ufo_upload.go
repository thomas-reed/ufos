package server

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/database"
	"github.com/thomas-reed/ufos/internal/objects"
)

func (s *Server) HandleUploadUFO(w http.ResponseWriter, r *http.Request) {
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
	ufoID := r.PathValue("uuid")

	obj, err := p.db.GetUFO(r.Context(), ufoID)
	if err != nil {
		respondWithError(
			w, http.StatusInternalServerError,
			"couldn't retrieve ufo from db",
			err,
		)
		return
	}
	// Check the status to make sure it's valid for upload
	status := objects.UFOStatus(obj.UploadStatus)
	if status == objects.StatusActive {
		respondWithError(
			w, http.StatusBadRequest,
			"cannot overwrite and active ufo",
			nil,
		)
		return
	}
	if status == objects.StatusUploading {
		respondWithError(
			w, http.StatusConflict,
			"upload already in progress",
			nil,
		)
		return
	}
	// Continue with upload in case of pending or failed states
	p.db.UpdateStatus(r.Context(), database.UpdateStatusParams{
		ID:           ufoID,
		UploadStatus: string(objects.StatusUploading),
	})

	// Open the physical file on the disk
	filePath := filepath.Join(p.RootFS, ufoID+".blob")
	file, err := os.Create(filePath) // os.Create truncates existing files (perfect for retries!)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "could not create file on disk", err)
		return
	}
	defer file.Close()

	// Write the file
	bytesWritten, err := io.Copy(file, r.Body)
	if err != nil {
		// If something fails, mark as failed so we can retry later
		p.db.UpdateStatus(r.Context(), database.UpdateStatusParams{
			ID:           ufoID,
			UploadStatus: string(objects.StatusFailed),
		})
		respondWithError(
			w,
			http.StatusInternalServerError,
			"stream interrupted",
			err,
		)
		return
	}

	// Verify expected filesize
	if bytesWritten != obj.SizeBytes {
		p.db.UpdateStatus(r.Context(), database.UpdateStatusParams{
			ID:           ufoID,
			UploadStatus: string(objects.StatusFailed),
		})
		respondWithError(
			w,
			http.StatusBadRequest,
			"uploaded size does not match metadata",
			nil,
		)
		return
	}

	// Mark object as active and send success
	res, _ := p.db.UpdateStatus(r.Context(), database.UpdateStatusParams{
		ID:           ufoID,
		UploadStatus: string(objects.StatusActive),
	})

	jsonResponse(
		w,
		http.StatusOK,
		api.UploadUFOResponse{
			ID:     ufoID,
			Status: objects.UFOStatus(res.UploadStatus),
		},
	)
}

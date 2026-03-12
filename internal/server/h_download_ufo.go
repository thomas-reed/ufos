package server

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/thomas-reed/ufos/internal/objects"
)

func (s *Server) HandleDownloadUFO(w http.ResponseWriter, r *http.Request) {
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

	obj, err := p.db.GetObject(r.Context(), ufoID)
	if err != nil {
		respondWithError(
			w, http.StatusNotFound,
			"couldn't retrieve object from db",
			err,
		)
		return
	}
	// Check the object to make sure it's downloadable
	if objects.ObjectStatus(obj.UploadStatus) != objects.StatusActive {
		respondWithError(
			w, http.StatusBadRequest,
			"cannot download non-active object",
			nil,
		)
		return
	}
	if obj.SizeBytes < 0 {
		respondWithError(
			w, http.StatusBadRequest,
			"cannot download folder objects",
			nil,
		)
		return
	}

	// Verify the file on disk
	filePath := filepath.Join(p.RootFS, ufoID+".blob")
	file, err := os.Open(filePath)
	if err != nil {
		respondWithError(
			w, http.StatusInternalServerError,
			"error opening file",
			err,
		)
		return
	}
	defer file.Close()
	// Construct headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", obj.SizeBytes))
	w.Header().Set("X-UFO-Metadata", base64.StdEncoding.EncodeToString(obj.Metadata))

	// Read file and stream
	_, err = io.Copy(w, file)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"stream interrupted",
			err,
		)
		return
	}
}

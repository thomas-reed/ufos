package server

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/database"
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

	ufo, err := p.db.GetUFO(r.Context(), ufoID)
	if err != nil {
		respondWithError(
			w, http.StatusNotFound,
			"couldn't retrieve ufo from db",
			err,
		)
		return
	}
	// Check the object to make sure it's downloadable
	if objects.UFOStatus(ufo.UploadStatus) != objects.StatusActive {
		respondWithError(
			w, http.StatusBadRequest,
			"cannot download non-active ufo",
			nil,
		)
		return
	}
	if ufo.SizeBytes < 0 {
		respondWithError(
			w, http.StatusBadRequest,
			"cannot download folder objects",
			nil,
		)
		return
	}

	// Verify/Open the file on disk
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
	w.Header().Set("Content-Length", fmt.Sprintf("%d", ufo.SizeBytes))
	if r.Header.Get(api.HeaderHost) == "" {
		w.Header().Set(api.HeaderMetadata, base64.StdEncoding.EncodeToString(ufo.Metadata))
	} else {
		wrappedKey, err := p.db.GetKeybyUFOIDAndPersonaID(
			r.Context(),
			database.GetKeybyUFOIDAndPersonaIDParams{
				UfoID:     ufoID,
				PersonaID: p.ID,
			},
		)
		if err != nil {
			respondWithError(
				w, http.StatusInternalServerError,
				"error getting wrapped key",
				err,
			)
			return
		}
		w.Header().Set(api.HeaderWrappedKey, base64.StdEncoding.EncodeToString(wrappedKey))
	}

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

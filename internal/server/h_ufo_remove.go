package server

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func (s *Server) HandleRemoveUFO(w http.ResponseWriter, r *http.Request) {
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

	deletedRes, err := p.db.DeleteUFO(r.Context(), ufoID)
	if err != nil {
		respondWithError(
			w, http.StatusNotFound,
			"couldn't find ufo in db",
			err,
		)
		return
	}

	if deletedRes.SizeBytes > 0 {
		filePath := filepath.Join(p.RootFS, ufoID+".blob")
		if err = os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			log.Printf("error deleting file: %s\n", err)
		}
	}

	jsonResponse(w, http.StatusNoContent, nil)
}

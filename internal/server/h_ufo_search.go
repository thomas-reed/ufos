package server

import (
	"net/http"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/objects"
)

func (s *Server) HandleSearch(w http.ResponseWriter, r *http.Request) {
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
	tagList := r.URL.Query()["tag"]
	if len(tagList) == 0 {
		respondWithError(
			w,
			http.StatusBadRequest,
			"no tags provided",
			nil,
		)
		return
	}

	ufoMap := make(map[string]api.UFO)
	for _, tag := range tagList {
		dbRows, err := p.db.GetUFOsByTag(r.Context(), tag)
		if err != nil {
			respondWithError(
				w,
				http.StatusInternalServerError,
				"db search error",
				err,
			)
			return
		}
		for _, row := range dbRows {
			ufoMap[row.ID] = api.UFO{
				ID:         row.ID,
				PrefixHash: row.PrefixHash,
				SizeBytes:  row.SizeBytes,
				Status:     objects.UFOStatus(row.UploadStatus),
				Metadata:   row.Metadata,
				CreatedAt:  row.CreatedAt,
				UpdatedAt:  row.UpdatedAt,
			}
		}
	}

	ufoList := make([]api.UFO, 0, len(ufoMap))
	for _, ufo := range ufoMap {
		ufoList = append(ufoList, ufo)
	}

	jsonResponse(w, http.StatusOK, ufoList)
}

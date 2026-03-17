package server

import (
	"net/http"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/objects"
)

func (s *Server) HandleList(w http.ResponseWriter, r *http.Request) {
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
	prefix := r.URL.Query().Get("prefix")
	if prefix == "" {
		respondWithError(
			w,
			http.StatusBadRequest,
			"no prefix provided",
			nil,
		)
		return
	}

	list, err := p.db.GetUFOsByParent(r.Context(), prefix)
	if err != nil {
		respondWithError(
			w,
			http.StatusInternalServerError,
			"db search error",
			err,
		)
		return
	}

	ufoList := make([]api.UFOItem, 0, len(list))
	for i := range list {
		ufoList = append(ufoList, api.UFOItem{
			ID:         list[i].ID,
			PrefixHash: list[i].PrefixHash,
			SizeBytes:  list[i].SizeBytes,
			Status:     objects.UFOStatus(list[i].UploadStatus),
			Metadata:   list[i].Metadata,
			CreatedAt:  list[i].CreatedAt,
			UpdatedAt:  list[i].UpdatedAt,
		})
	}

	jsonResponse(w, http.StatusOK, ufoList)
}

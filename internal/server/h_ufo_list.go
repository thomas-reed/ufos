package server

import (
	"net/http"

	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/database"
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
	tagList := r.URL.Query()["tag"]
	if prefix == "" && len(tagList) == 0 {
		respondWithError(
			w,
			http.StatusBadRequest,
			"no prefix or tags provided",
			nil,
		)
		return
	}

	var list []database.Ufo
	var err error

	if prefix != "" && len(tagList) == 0 { // Prefix-only
		list, err = p.db.GetUFOsByParent(r.Context(), prefix)
		if err != nil {
			respondWithError(
				w,
				http.StatusInternalServerError,
				"db search error",
				err,
			)
			return
		}
	} else if prefix == "" && len(tagList) > 0 { // Tags-only
		list, err = p.db.GetUFOsByTags(r.Context(), database.GetUFOsByTagsParams{
			Tags:     tagList,
			TagCount: int64(len(tagList)),
		})
	} else { // Prefix and Tags
		tagList = append(tagList, prefix)
		list, err = p.db.GetUFOsByTags(r.Context(), database.GetUFOsByTagsParams{
			Tags:     append(tagList),
			TagCount: int64(len(tagList)),
		})
	}

	ufoList := make([]api.UFO, 0, len(list))
	for i := range list {
		ufoList = append(ufoList, api.UFO{
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

package server

import (
	"net/http"

	"github.com/thomas-reed/ufos/internal/api"
)

func HandleHealth(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, api.EmptyResponse{})
}

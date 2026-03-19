package server

import (
	"net/http"

	"github.com/thomas-reed/ufos/internal/api"
)

func (s *Server) Router() *http.ServeMux {
	authHash := s.Authenticate(true)
	authNoHash := s.Authenticate(false)

	mux := http.NewServeMux()

	// PUBLIC BRANCH
	mux.HandleFunc("GET /healthz", HandleHealth)

	// PROTECTED BRANCH
	// Auth requires body hash
	mux.Handle("POST "+api.RouteInit, authHash(http.HandlerFunc(s.HandleInitPersona)))
	mux.Handle("GET "+api.RouteUFOs, authHash(http.HandlerFunc(s.HandleList)))
	mux.Handle("POST "+api.RouteUFOs, authHash(http.HandlerFunc(s.HandleCreateUFO)))
	mux.Handle("GET "+api.RouteUFOs+"/{uuid}", authHash(http.HandlerFunc(s.HandleDownloadUFO)))
	mux.Handle("PATCH "+api.RouteUFOs+"/{uuid}", authHash(http.HandlerFunc(s.HandleUpdateUFO)))
	mux.Handle("DELETE "+api.RouteUFOs+"/{uuid}", authHash(http.HandlerFunc(s.HandleRemoveUFO)))
	mux.Handle("GET "+api.RouteSearch, authHash(http.HandlerFunc(s.HandleSearch)))
	mux.Handle("POST "+api.RouteOrbit, authHash(http.HandlerFunc(s.HandleAddToOrbit)))
	mux.Handle("GET "+api.RouteOrbit, authHash(http.HandlerFunc(s.HandleOrbitList)))
	mux.Handle("DELETE "+api.RouteOrbit+"/{id}", authHash(http.HandlerFunc(s.HandleRemoveFromOrbit)))
	// Auth does not require body hash
	mux.Handle("PUT "+api.RouteUFOs+"/{uuid}", authNoHash(http.HandlerFunc(s.HandleUploadUFO)))

	// "NEW PERSONA" BRANCH (request must contain valid NEW_PERSONA_TOKEN)
	mux.HandleFunc("POST "+api.RouteRegister, s.HandleCreatePersona)

	return mux
}

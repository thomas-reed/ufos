package server

import (
	"net/http"

	"github.com/thomas-reed/ufos/internal/api"
)

func (s *Server) Router() *http.ServeMux {
	mux := http.NewServeMux()

	// PUBLIC BRANCH
	mux.HandleFunc("GET /healthz", HandleHealth)
	// Protected by Registration token
	mux.HandleFunc("POST "+api.RouteRegister, s.HandleCreatePersona)

	// PROTECTED BRANCH
	// Handle json requests
	mux.Handle("POST "+api.RouteInit, s.Authenticate(http.HandlerFunc(s.HandleInitPersona)))
	mux.Handle("GET "+api.RouteUFOs, s.Authenticate(http.HandlerFunc(s.HandleList)))
	mux.Handle("POST "+api.RouteUFOs, s.Authenticate(http.HandlerFunc(s.HandleCreateUFO)))
	mux.Handle("PATCH "+api.RouteUFOs+"/{uuid}", s.Authenticate(http.HandlerFunc(s.HandleUpdateUFO)))
	mux.Handle("DELETE "+api.RouteUFOs+"/{uuid}", s.Authenticate(http.HandlerFunc(s.HandleRemoveUFO)))
	mux.Handle("GET "+api.RouteSearch, s.Authenticate(http.HandlerFunc(s.HandleSearch)))
	mux.Handle("POST "+api.RouteOrbit, s.Authenticate(http.HandlerFunc(s.HandleAddToOrbit)))
	mux.Handle("GET "+api.RouteOrbit, s.Authenticate(http.HandlerFunc(s.HandleOrbitList)))
	mux.Handle("DELETE "+api.RouteOrbit+"/{id}", s.Authenticate(http.HandlerFunc(s.HandleRemoveFromOrbit)))
	// Handle streaming requests
	mux.Handle("PUT "+api.RouteUFOs+"/{uuid}", s.AuthenticateStream(http.HandlerFunc(s.HandleUploadUFO)))
	mux.Handle("GET "+api.RouteUFOs+"/{uuid}", s.AuthenticateStream(http.HandlerFunc(s.HandleDownloadUFO)))

	return mux
}

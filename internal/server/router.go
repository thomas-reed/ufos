package server

import (
	"net/http"

	"github.com/thomas-reed/ufos/internal/api"
)

func (s *Server) Router() *http.ServeMux {
	mux := http.NewServeMux()

	// PUBLIC BRANCH
	mux.HandleFunc("GET /healthz", HandleHealth)
	mux.HandleFunc("GET "+api.RoutePersonas+"/{id}", s.HandleGetPersonaKeys)
	// Protected by Registration token
	mux.HandleFunc("POST "+api.RoutePersonas, s.HandleCreatePersona)

	// PROTECTED BRANCH
	mux.Handle("POST "+api.RouteInit, s.Authenticate(http.HandlerFunc(s.HandleInitPersona)))
	mux.Handle("GET "+api.RouteUFOs, s.Authenticate(http.HandlerFunc(s.HandleList)))
	mux.Handle("POST "+api.RouteUFOs, s.Authenticate(http.HandlerFunc(s.HandleCreateUFO)))
	mux.Handle("HEAD "+api.RouteUFOs+"/{uuid}", s.Authenticate(http.HandlerFunc(s.HandleGetUFOMetadata)))
	mux.Handle("PATCH "+api.RouteUFOs+"/{uuid}", s.Authenticate(http.HandlerFunc(s.HandleUpdateUFO)))
	mux.Handle("DELETE "+api.RouteUFOs+"/{uuid}", s.Authenticate(http.HandlerFunc(s.HandleRemoveUFO)))
	mux.Handle("POST "+api.RouteOrbit, s.Authenticate(http.HandlerFunc(s.HandleAddToOrbit)))
	mux.Handle("GET "+api.RouteOrbit, s.Authenticate(http.HandlerFunc(s.HandleOrbitList)))
	mux.Handle("GET "+api.RouteOrbit+"/{id}", s.Authenticate(http.HandlerFunc(s.HandleGetFromOrbit)))
	mux.Handle("DELETE "+api.RouteOrbit+"/{id}", s.Authenticate(http.HandlerFunc(s.HandleRemoveFromOrbit)))

	// Handle streaming requests
	mux.Handle("PUT "+api.RouteUFOs+"/{uuid}", s.AuthenticateStream(http.HandlerFunc(s.HandleUploadUFO)))
	mux.Handle("GET "+api.RouteUFOs+"/{uuid}", s.AuthenticateStream(http.HandlerFunc(s.HandleDownloadUFO)))

	return mux
}

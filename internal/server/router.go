package server

import "net/http"

func (s *Server) Router() *http.ServeMux {
	authHash := s.Authenticate(true)
	authNoHash := s.Authenticate(false)

	mux := http.NewServeMux()

	// PUBLIC BRANCH
	mux.HandleFunc("GET /healthz", HandleHealth)

	// PROTECTED BRANCH
	// Auth requires body hash
	mux.Handle("POST /api/init", authHash(http.HandlerFunc(s.HandleInitPersona)))
	mux.Handle("GET /api/ufos", authHash(http.HandlerFunc(s.HandleList)))
	mux.Handle("POST /api/ufos", authHash(http.HandlerFunc(s.HandleCreateUFO)))
	mux.Handle("GET /api/ufos/{uuid}", authHash(http.HandlerFunc(s.HandleDownloadUFO)))
	mux.Handle("PATCH /api/ufos/{uuid}", authHash(http.HandlerFunc(s.HandleUpdateUFO)))
	mux.Handle("DELETE /api/ufos/{uuid}", authHash(http.HandlerFunc(s.HandleRemoveUFO)))
	mux.Handle("GET /api/tags", authHash(http.HandlerFunc(s.HandleSearch)))
	mux.Handle("POST /api/orbit", authHash(http.HandlerFunc(s.HandleAddToOrbit)))
	mux.Handle("GET /api/orbit", authHash(http.HandlerFunc(s.HandleOrbitList)))
	mux.Handle("DELETE /api/orbit/{id}", authHash(http.HandlerFunc(s.HandleRemoveFromOrbit)))
	// Auth does not require body hash
	mux.Handle("PUT /api/ufos/{uuid}", authNoHash(http.HandlerFunc(s.HandleUploadUFO)))

	// "NEW PERSONA" BRANCH (request must contain valid NEW_PERSONA_TOKEN)
	mux.HandleFunc("POST /api/personas", s.HandleCreatePersona)

	return mux
}

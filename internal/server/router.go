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
	mux.Handle("GET /api/objects", authHash(http.HandlerFunc(s.HandleList)))
	mux.Handle("GET /api/tags", authHash(http.HandlerFunc(s.HandleSearch)))
	mux.Handle("POST /api/objects", authHash(http.HandlerFunc(s.HandleCreateUFO)))
	mux.Handle("GET /api/objects/{uuid}", authHash(http.HandlerFunc(s.HandleDownloadUFO)))
	mux.Handle("PATCH /api/objects/{uuid}", authHash(http.HandlerFunc(s.HandleUpdateUFO)))
	mux.Handle("DELETE /api/objects/{uuid}", authHash(http.HandlerFunc(s.HandleRemoveUFO)))
	mux.Handle("POST /api/init", authHash(http.HandlerFunc(s.HandleInitPersona)))
	// Auth does not require body hash
	mux.Handle("PUT /api/objects/{uuid}", authNoHash(http.HandlerFunc(s.HandleUploadObject)))

	// "NEW PERSONA" BRANCH (request must contain valid NEW_PERSONA_TOKEN)
	mux.HandleFunc("POST /api/personas", s.HandleCreatePersona)

	return mux
}

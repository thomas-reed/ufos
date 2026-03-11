package server

import "net/http"

func (s *Server) Router() *http.ServeMux {
	mainMux := http.NewServeMux()

	// PUBLIC BRANCH
	mainMux.HandleFunc("GET /healthz", HandleHealth)

	// PROTECTED BRANCH (requests go through Authenticate method)
	apiMux := http.NewServeMux()

	// Gets all UFO metadata by prefix_hash
	apiMux.HandleFunc("GET /api/objects", s.HandleList)

	// Search all UFOs by tag_hashes
	apiMux.HandleFunc("GET /api/tags", s.HandleSearch)

	// Upload
	// Step 1: Create the UFO in database (metadata)
	apiMux.HandleFunc("POST /api/objects", s.HandleCreateUFO)
	// Step 2: Stream the actual file bytes to the disk
	apiMux.HandleFunc("PUT /api/objects/{uuid}", s.HandleUploadObject)

	// Download UFO
	apiMux.HandleFunc("GET /api/objects/{uuid}", s.HandleDownload)

	// Update UFO metadata
	apiMux.HandleFunc("PATCH /api/objects/{uuid}", s.HandleUpdateUFO)

	// Remove UFO
	apiMux.HandleFunc("DELETE /api/objects/{uuid}", s.HandleRemove)

	// Authenticated endpoint to generate a NEW_PERSONA_TOKEN
	apiMux.HandleFunc("POST /api/init", s.HandleInitPersona)

	// Mount the apiMux under the Authenticate middleware
	mainMux.Handle("/api/", s.Authenticate(apiMux))

	// "NEW PERSONA" BRANCH (request must contain valid NEW_PERSONA_TOKEN)
	mainMux.HandleFunc("POST /api/personas", s.HandleCreatePersona)

	return mainMux
}

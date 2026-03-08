package server

import "net/http"

func (s *Server) RouteSetup() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/healthz", healthHandler)
	mux.HandleFunc("POST /api/upload", s.uploadHandler)
	mux.HandleFunc("POST /api/list", s.listHandler)
	mux.HandleFunc("PUT /api/download", s.downloadHandler)
	return mux
}

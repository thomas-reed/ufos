package server

import (
	"net/http"
)

func (s *Server) HandleUploadObject(w http.ResponseWriter, r *http.Request) {
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
	ufoID := r.PathValue("id")

	// 1. Check if the object exists and is 'pending'
	obj, err := p.db.GetObject(r.Context(), ufoID)
	if err != nil {
			// ... handle not found ...
	}
	
	// 2. Update status to 'uploading'
	p.db.UpdateStatus(r.Context(), database.UpdateStatusParams{
			ID: ufoID,
			UploadStatus: string(objects.StatusUploading),
	})

	// 3. Create the file on disk
	// (Path: p.RootFS + "/" + ufoID + ".blob")
	
	// 4. Use io.Copy to stream r.Body into the file
	
	// 5. Success -> Update status to 'active'
	// 6. Failure -> Update status to 'failed'
}

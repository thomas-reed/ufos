package server

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/thomas-reed/ufos/internal/api"
	"github.com/thomas-reed/ufos/internal/database"
)

func (s *Server) HandleCreatePersona(w http.ResponseWriter, r *http.Request) {
	token := r.Header.Get(api.HeaderRegistration)

	s.mu.RLock()
	validToken := s.registrationToken != "" && s.registrationToken == token
	s.mu.RUnlock()

	if !validToken {
		respondWithError(w, http.StatusUnauthorized, "", nil)
		return
	}

	var req api.NewPersonaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "invalid json body", err)
		return
	}

	s.mu.RLock()
	for _, p := range s.Registry {
		if req.ID == p.ID {
			respondWithError(w, http.StatusConflict, "duplicate ID found", nil)
			return
		}
	}
	s.mu.RUnlock()

	var err error
	rootFS := filepath.Join(s.dataPath, req.ID, "objects")
	if err = os.MkdirAll(rootFS, 0700); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error creating persona FS", err)
		return
	}
	dbPath := filepath.Join(s.dataPath, req.ID, "ufos.db")
	dbConn, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error opening DB", err)
		return
	}
	if err := goose.Up(dbConn, "sql/schema"); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Error migrating DB", err)
		return
	}
	err = s.AddPersona(Persona{
		ID:          req.ID,
		SigningKey:  req.SigningKey,
		ExchangeKey: req.ExchangeKey,
		RootFS:      rootFS,
		DBPath:      dbPath,
		DBConn:      dbConn,
		db:          database.New(dbConn),
	})
	if err != nil {
		respondWithError(w, http.StatusConflict, "duplicate ID found", nil)
		return
	}

	jsonResponse(w, http.StatusCreated, api.CreatePersonaResponse{
		ID: req.ID,
	})

	log.Printf("Persona %s created successfully\n", req.ID)
}

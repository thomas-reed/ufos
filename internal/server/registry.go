package server

import (
	"crypto/ed25519"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/thomas-reed/ufos/internal/database"
)

type Persona struct {
	ID        string            `json:"id"`           // Persona ID
	PublicKey ed25519.PublicKey `json:"public_key"`   // ED25519 public key
	RootFS    string            `json:"root_fs_path"` // root directory for the persona's file store
	DbPath    string            `json:"db_path"`      // server local path to the sqlite db file
	dbConn    *sql.DB
	db        *database.Queries
}

func (s *Server) LoadRegistry(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("Error reading registry file")
	}
	if err := json.Unmarshal(data, &s.registry); err != nil {
		return fmt.Errorf("Error unmarshalling registry data")
	}
	return nil
}

func (s *Server) SaveRegistry() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(s.registry, "", "  ")
	if err != nil {
		return err
	}
	// 0600 ensures only the server process can read/write this file
	return os.WriteFile(s.registryFilepath, data, 0600)
}

func (s *Server) GetPersona(personaID string) (*Persona, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.registry[personaID]
	return p, ok
}

func (s *Server) AddPersona(p Persona) error {
	s.mu.Lock()
	newP := p
	s.registry[p.ID] = &newP
	s.mu.Unlock()

	return s.SaveRegistry()
}

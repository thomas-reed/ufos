package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/thomas-reed/ufos/internal/database"
)

type Persona struct {
	ID          string `json:"id"`           // Persona ID
	SigningKey  []byte `json:"signing_key"`  // ED25519 public key for signature verification
	ExchangeKey []byte `json:"exchange_key"` // X25519 public key for encryption
	RootFS      string `json:"root_fs_path"` // root directory for the persona's file store
	DBPath      string `json:"db_path"`      // server local path to the sqlite db file
	DBConn      *sql.DB
	db          *database.Queries
}

func (s *Server) LoadRegistry() error {
	data, err := os.ReadFile(s.registryPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			s.Registry = make(map[string]*Persona)
			return s.SaveRegistry()
		}
		return fmt.Errorf("Error reading registry file: %w", err)
	}
	if err := json.Unmarshal(data, &s.Registry); err != nil {
		return fmt.Errorf("Error unmarshalling registry data: %w", err)
	}
	return nil
}

func (s *Server) SaveRegistry() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := json.MarshalIndent(s.Registry, "", "  ")
	if err != nil {
		return err
	}
	// 0600 ensures only the server process can read/write this file
	return os.WriteFile(s.registryPath, data, 0600)
}

func (s *Server) GetPersona(personaID string) (*Persona, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.Registry[personaID]
	return p, ok
}

func (s *Server) AddPersona(p Persona) error {
	s.mu.Lock()
	for _, persona := range s.Registry {
		if p.ID == persona.ID {
			return fmt.Errorf("Duplicate ID")
		}
	}
	newP := p
	s.Registry[p.ID] = &newP
	s.registrationToken = "" // Ensure token is wiped out
	s.mu.Unlock()

	return s.SaveRegistry()
}

package server

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/thomas-reed/ufos/internal/database"
)

type Server struct {
	HTTPServer       *http.Server
	Port             string
	newPersonaToken  string
	mu               sync.RWMutex
	registryFilepath string
	Registry         map[string]*Persona
}

func NewServer() (*Server, error) {
	s := Server{}

	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("Could not load env file: %w", err)
	}
	s.Port = os.Getenv("PORT")
	if s.Port == "" {
		return nil, fmt.Errorf("Port not found")
	}
	s.registryFilepath = os.Getenv("REGISTRY_FILEPATH")
	if s.registryFilepath == "" {
		return nil, fmt.Errorf("Registry filepath not found")
	}

	err = s.LoadRegistry(s.registryFilepath)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, fmt.Errorf("Error running goose SetDialect: %w", err)
	}

	for _, persona := range s.Registry {
		dbConn, err := sql.Open("sqlite3", persona.DBPath)
		if err != nil {
			return nil, fmt.Errorf("Could not open db for %s: %w", persona.ID, err)
		}

		if err := goose.Up(dbConn, "sql/schema"); err != nil {
			return nil, fmt.Errorf("Migration failed for %s: %w", persona.ID, err)
		}
		persona.DBConn = dbConn
		persona.db = database.New(dbConn)
	}

	s.HTTPServer = &http.Server{
		Addr:    ":" + s.Port,
		Handler: s.Router(),
	}

	return &s, nil
}

package server

import (
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/pressly/goose/v3"
	"github.com/thomas-reed/ufos/internal/crypto"
	"github.com/thomas-reed/ufos/internal/database"
)

const (
	tokenLifetime   = 15 * time.Minute
	janitorInterval = 5 * time.Minute
)

type Server struct {
	HTTPServer        *http.Server
	Port              string
	mu                sync.RWMutex
	WG                sync.WaitGroup
	registrationToken string
	tokenCreated      time.Time
	dataPath          string
	registryPath      string
	Registry          map[string]*Persona
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
	s.dataPath = os.Getenv("DATA_PATH")
	if s.dataPath == "" {
		return nil, fmt.Errorf("Data path not found")
	}

	s.registryPath = filepath.Join(s.dataPath, "registry.json")
	err = s.LoadRegistry()
	if err != nil {
		return nil, fmt.Errorf("Registry could not be loaded: %w", err)
	}

	if len(s.Registry) == 0 {
		// Check if the admin provided a seed token
		token := os.Getenv("UFO_BOOTSTRAP_TOKEN")
		if token != "" {
			s.registrationToken = token
			log.Println("Bootstrap token loaded from environment.")
		} else {
			// Fallback: Generate one and PRINT IT
			token, err := crypto.GenerateKey()
			if err != nil {
				log.Fatalf("Error generating registration token: %s", err)
			}
			s.registrationToken = base64.StdEncoding.EncodeToString(token)
			s.tokenCreated = time.Time{}
			log.Printf("!!! INITIAL BOOTSTRAP TOKEN: %s", s.registrationToken)
		}
	}

	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, fmt.Errorf("Error running goose SetDialect: %w", err)
	}

	for _, persona := range s.Registry {
		dbConn, err := sql.Open("sqlite3", persona.DBPath+"?_foreign_keys=on")
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

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
	schemaDir         string
	mode              string
	tlsEnabled        bool
	tls               tlsInfo
	registrationToken string
	tokenCreated      time.Time
	dataPath          string
	registryPath      string
	Registry          map[string]*Persona
}

type tlsInfo struct {
	certDir  string
	certFile string
	keyFile  string
}

func NewServer() (*Server, error) {
	s := Server{}
	s.schemaDir = "/app/sql/schema"

	_ = godotenv.Load()

	s.Port = os.Getenv("PORT")
	if s.Port == "" {
		return nil, fmt.Errorf("Port not found")
	}
	s.dataPath = os.Getenv("DATA_PATH")
	if s.dataPath == "" {
		return nil, fmt.Errorf("Data path not found")
	}
	s.mode = os.Getenv("MODE")
	if s.mode == "" {
		s.mode = "dev"
	}

	if err := ensureDir(s.dataPath); err != nil {
		return nil, err
	}

	s.registryPath = filepath.Join(s.dataPath, "registry.json")
	if err := s.LoadRegistry(); err != nil {
		return nil, err
	}

	if len(s.Registry) == 0 {
		// Check if the admin provided a seed token
		token := os.Getenv("UFO_BOOTSTRAP_TOKEN")
		if token != "" {
			s.registrationToken = token
			log.Println("Bootstrap token loaded from environment.")
		} else {
			// Fallback: Generate one and print it to console
			token, err := crypto.GenerateKey()
			if err != nil {
				return nil, fmt.Errorf("Error generating registration token: %w", err)
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

		if err := goose.Up(dbConn, s.schemaDir); err != nil {
			return nil, fmt.Errorf("Migration failed for %s: %w", persona.ID, err)
		}
		persona.DBConn = dbConn
		persona.db = database.New(dbConn)
	}

	if s.mode == "prod" {
		if err := s.initTLS(); err != nil {
			return nil, err
		}
	}
	s.HTTPServer = &http.Server{
		Addr:    ":" + s.Port,
		Handler: s.Router(),
	}

	return &s, nil
}

func ensureDir(dir string) error {
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("Data path couldn't be found/created: %w", err)
	}
	return nil
}

func (s *Server) initTLS() error {
	s.tls.certDir = filepath.Join(s.dataPath, "certs")
	if err := ensureDir(s.tls.certDir); err != nil {
		return err
	}
	// TODO: ACME LETSENCRYPT STUFF
	s.tlsEnabled = true
	return nil
}

func (s *Server) StartServer() error {
	log.Printf("Server listening on port: %s\n", s.Port)
	if s.tlsEnabled {
		if err := s.HTTPServer.ListenAndServeTLS(s.tls.certFile, s.tls.keyFile); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("Server start error: %w", err)
		}
	} else {
		if err := s.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			return fmt.Errorf("Server start error: %w", err)
		}
	}
	return nil
}

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
	registry         map[string]*Persona
}

func NewServer() (*Server, error) {
	server := Server{}

	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("Could not load env file: %w", err)
	}
	server.Port = os.Getenv("PORT")
	if server.Port == "" {
		return nil, fmt.Errorf("Port not found")
	}
	server.registryFilepath = os.Getenv("REGISTRY_FILEPATH")
	if server.registryFilepath == "" {
		return nil, fmt.Errorf("Registry filepath not found")
	}
	server.newPersonaToken = os.Getenv("NEW_PERSONA_TOKEN")

	err = server.LoadRegistry(server.registryFilepath)

	if err := goose.SetDialect("sqlite3"); err != nil {
		return nil, fmt.Errorf("Error running goose SetDialect: %w", err)
	}

	for _, persona := range server.registry {
		dbConn, err := sql.Open("sqlite3", persona.DbPath)
		if err != nil {
			return nil, fmt.Errorf("Could not open database connection: %w", err)
		}

		if err := goose.Up(persona.dbConn, "sql/schema"); err != nil {
			return nil, fmt.Errorf("Error running goose Up for %s: %w", persona.ID, err)
		}
		persona.dbConn = dbConn
		persona.db = database.New(dbConn)
	}

	server.HTTPServer = &http.Server{
		Addr:    ":" + server.Port,
		Handler: server.RouteSetup(),
	}

	return &server, nil
}

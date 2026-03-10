package main

import (
	"log"

	"github.com/thomas-reed/ufos/internal/server"
)

func main() {
	s, err := server.NewServer()
	for id := range s.Registry {
		defer s.Registry[id].DBConn.Close()
	}

	if err != nil {
		log.Fatalf("Error initializing server: %v", err)
	}

	log.Printf("Server listening on port: %s\n", s.Port)
	log.Fatal(s.HTTPServer.ListenAndServe())
}

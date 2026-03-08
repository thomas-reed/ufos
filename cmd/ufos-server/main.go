package main

import (
	"log"

	"github.com/thomas-reed/ufos/internal/server"
)

func main() {
	svr, err := server.NewServer()
	if err != nil {
		log.Fatalf("Error initializing server: %v", err)
	}

	log.Printf("Server listening on port: %s\n", svr.Port)
	log.Fatal(svr.HTTPServer.ListenAndServe())
}

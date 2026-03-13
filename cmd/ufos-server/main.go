package main

import (
	"context"
	"net/http"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/thomas-reed/ufos/internal/server"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v. Initiating shutdown...", sig)
		cancel()
	}()

	s, err := server.NewServer()
	if err != nil {
		log.Fatalf("Error initializing server: %v", err)
	}

	s.WG.Add(1)
	go s.StartJanitor(ctx)

	go func() {
    log.Printf("Server listening on port: %s\n", s.Port)
    if err := s.HTTPServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatalf("Server error: %v", err)
    }
	}()
	
	// WAIT for the signal to shutdown
	<-ctx.Done()

	log.Println("Shutting down HTTP server...")
	// Give the server 10 seconds to finish existing requests
	shutdownCtx, shutdownCancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer shutdownCancel()

	if err := s.HTTPServer.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("Graceful shutdown failed: %v", err)
	}
	
	log.Println("Waiting for background tasks to finish...")
	s.WG.Wait()

	for id := range s.Registry {
		s.Registry[id].DBConn.Close()
	}
	log.Println("Server exited cleanly.")
}
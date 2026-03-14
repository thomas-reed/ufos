package server

import (
	"context"
	"log"
	"time"
)

func (s *Server) StartJanitor(ctx context.Context) {
	defer s.WG.Done()
	ticker := time.NewTicker(janitorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Janitor: Starting scheduled sweep...")
			// Clean up expired registration token
			s.mu.Lock()
			// Check the registration token is populated, has expired, and is not the bootstrap token
			if s.registrationToken != "" && time.Since(s.tokenCreated) > tokenLifetime && !s.tokenCreated.IsZero(){
				s.registrationToken = ""
				s.tokenCreated = time.Time{}
				log.Println("Janitor: Cleared expired registration token.")
			}
			s.mu.Unlock()
			
			// Clean up requests tables in each persona db
			s.mu.RLock()
			for _, p := range s.Registry {
				if err := p.db.DeleteStaleRequests(ctx); err != nil {
					log.Printf("Janitor: Error deleting stale requests from %s:\n %s\n", p.ID, err)
				}
			}
			s.mu.RUnlock()
			log.Println("Janitor: Sweep completed.")
		case <-ctx.Done():
			log.Println("Janitor: Shutting down...")
			return
		}
	}
}
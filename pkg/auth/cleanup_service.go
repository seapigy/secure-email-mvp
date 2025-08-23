package auth

import (
	"log"
	"time"
)

// CleanupService handles periodic cleanup of expired pending signups
type CleanupService struct {
	pendingSignupService *PendingSignupService
	interval             time.Duration
	stopChan             chan struct{}
}

// NewCleanupService creates a new cleanup service
func NewCleanupService(pendingSignupService *PendingSignupService, interval time.Duration) *CleanupService {
	return &CleanupService{
		pendingSignupService: pendingSignupService,
		interval:             interval,
		stopChan:             make(chan struct{}),
	}
}

// Start begins the periodic cleanup process
func (cs *CleanupService) Start() {
	log.Printf("Starting cleanup service with interval: %v", cs.interval)

	ticker := time.NewTicker(cs.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cs.runCleanup()
		case <-cs.stopChan:
			log.Printf("Cleanup service stopped")
			return
		}
	}
}

// Stop stops the cleanup service
func (cs *CleanupService) Stop() {
	close(cs.stopChan)
}

// runCleanup performs the actual cleanup of expired pending signups
func (cs *CleanupService) runCleanup() {
	// Clean up expired pending signups
	cleanedCount, err := cs.pendingSignupService.CleanupExpiredSignups()
	if err != nil {
		log.Printf("Failed to cleanup expired pending signups: %v", err)
		return
	}

	if cleanedCount > 0 {
		log.Printf("Cleaned up %d expired pending signups", cleanedCount)
	}
}

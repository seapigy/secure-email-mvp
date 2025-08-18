package audit

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"time"
)

// ConcurrentAccessManager handles protection against concurrent access to the same email
// Enhanced for Micro-Iteration 4.22: Prevent multiple simultaneous retrievals of the same email blob
// Use a short-lived lock (e.g., 2 seconds) to prevent race conditions that could bypass
// burn_after_read or attempt limits
type ConcurrentAccessManager struct {
	db     *sql.DB
	locks  map[string]*emailLock
	lockMu sync.RWMutex
}

// emailLock represents a lock for a specific email with timeout functionality
type emailLock struct {
	mu       sync.Mutex
	lockedAt time.Time
	timeout  time.Duration
}

// NewConcurrentAccessManager creates a new concurrent access manager
func NewConcurrentAccessManager(db *sql.DB) *ConcurrentAccessManager {
	return &ConcurrentAccessManager{
		db:    db,
		locks: make(map[string]*emailLock),
	}
}

// AcquireLock attempts to acquire a lock for the given email ID
// Returns true if lock was acquired, false if already locked
// Enhanced for Micro-Iteration 4.22: Short-lived lock (2 seconds) to prevent race conditions
func (m *ConcurrentAccessManager) AcquireLock(emailID string) bool {
	m.lockMu.Lock()
	defer m.lockMu.Unlock()

	// Get or create lock for this email
	lock, exists := m.locks[emailID]
	if !exists {
		lock = &emailLock{
			timeout: 2 * time.Second, // Short-lived lock as per Micro-Iteration 4.22
		}
		m.locks[emailID] = lock
	}

	// Check if lock has expired
	if exists && time.Since(lock.lockedAt) > lock.timeout {
		log.Printf("Concurrent access lock expired for email: %s, allowing new access", emailID)
		// Reset the lock
		lock.lockedAt = time.Time{}
	}

	// Try to acquire the lock
	acquired := lock.mu.TryLock()
	if acquired {
		lock.lockedAt = time.Now()
		log.Printf("Concurrent access lock acquired for email: %s (timeout: %v)", emailID, lock.timeout)
	} else {
		log.Printf("Concurrent access lock already held for email: %s (held for: %v)",
			emailID, time.Since(lock.lockedAt))
	}

	return acquired
}

// ReleaseLock releases the lock for the given email ID
func (m *ConcurrentAccessManager) ReleaseLock(emailID string) {
	m.lockMu.RLock()
	lock, exists := m.locks[emailID]
	m.lockMu.RUnlock()

	if exists {
		lock.mu.Unlock()
		log.Printf("Concurrent access lock released for email: %s", emailID)
	}
}

// WaitForLock waits for a lock to become available with a timeout
// Enhanced for Micro-Iteration 4.22: Prevents race conditions that could bypass security controls
func (m *ConcurrentAccessManager) WaitForLock(ctx context.Context, emailID string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if m.AcquireLock(emailID) {
				return nil
			}
			time.Sleep(100 * time.Millisecond) // Wait 100ms before retrying
		}
	}

	return fmt.Errorf("timeout waiting for lock on email %s", emailID)
}

// IsLocked checks if an email is currently locked without acquiring the lock
func (m *ConcurrentAccessManager) IsLocked(emailID string) bool {
	m.lockMu.RLock()
	defer m.lockMu.RUnlock()

	lock, exists := m.locks[emailID]
	if !exists {
		return false
	}

	// Check if lock has expired
	if time.Since(lock.lockedAt) > lock.timeout {
		return false
	}

	// Try to acquire lock to check if it's available
	acquired := lock.mu.TryLock()
	if acquired {
		lock.mu.Unlock()
		return false
	}
	return true
}

// GetLockStatus returns the current lock status for an email
func (m *ConcurrentAccessManager) GetLockStatus(emailID string) map[string]interface{} {
	m.lockMu.RLock()
	defer m.lockMu.RUnlock()

	lock, exists := m.locks[emailID]
	if !exists {
		return map[string]interface{}{
			"locked":     false,
			"email_id":   emailID,
			"held_since": nil,
			"timeout":    nil,
		}
	}

	status := map[string]interface{}{
		"locked":   true,
		"email_id": emailID,
		"timeout":  lock.timeout.String(),
	}

	if !lock.lockedAt.IsZero() {
		status["held_since"] = lock.lockedAt
		status["time_remaining"] = lock.timeout - time.Since(lock.lockedAt)
	}

	return status
}

// CleanupLocks removes locks for emails that are no longer being accessed
// Enhanced for Micro-Iteration 4.22: Automatic cleanup of expired locks
func (m *ConcurrentAccessManager) CleanupLocks() {
	m.lockMu.Lock()
	defer m.lockMu.Unlock()

	now := time.Now()
	cleanedCount := 0

	for emailID, lock := range m.locks {
		// Remove locks that have been expired for more than 5 minutes
		if now.Sub(lock.lockedAt) > lock.timeout+5*time.Minute {
			delete(m.locks, emailID)
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		log.Printf("Cleaned up %d expired concurrent access locks", cleanedCount)
	}
}

// GetActiveLocks returns information about all currently active locks
// Enhanced for Micro-Iteration 4.22: Admin debugging capabilities
func (m *ConcurrentAccessManager) GetActiveLocks() map[string]interface{} {
	m.lockMu.RLock()
	defer m.lockMu.RUnlock()

	activeLocks := make(map[string]interface{})
	now := time.Now()

	for emailID, lock := range m.locks {
		// Only include locks that are actually active (not expired)
		if now.Sub(lock.lockedAt) <= lock.timeout {
			activeLocks[emailID] = map[string]interface{}{
				"locked_at":      lock.lockedAt,
				"timeout":        lock.timeout.String(),
				"time_remaining": lock.timeout - now.Sub(lock.lockedAt),
			}
		}
	}

	return map[string]interface{}{
		"active_locks": activeLocks,
		"total_locks":  len(activeLocks),
	}
}

// ForceReleaseLock forcefully releases a lock for an email (admin function)
// Enhanced for Micro-Iteration 4.22: Admin debugging capabilities
func (m *ConcurrentAccessManager) ForceReleaseLock(emailID string) bool {
	m.lockMu.Lock()
	defer m.lockMu.Unlock()

	lock, exists := m.locks[emailID]
	if !exists {
		return false
	}

	// Try to acquire the lock to release it
	acquired := lock.mu.TryLock()
	if acquired {
		lock.mu.Unlock()
		delete(m.locks, emailID)
		log.Printf("Forcefully released concurrent access lock for email: %s", emailID)
		return true
	}

	// If we can't acquire it, it might be held by another goroutine
	// We'll mark it for cleanup by setting a very old timestamp
	lock.lockedAt = time.Now().Add(-10 * time.Minute)
	log.Printf("Marked concurrent access lock for email %s for cleanup", emailID)
	return false
}

// StartCleanupLoop starts a background goroutine to periodically cleanup expired locks
// Enhanced for Micro-Iteration 4.22: Automatic maintenance
func (m *ConcurrentAccessManager) StartCleanupLoop(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Printf("Stopping concurrent access lock cleanup loop")
				return
			case <-ticker.C:
				m.CleanupLocks()
			}
		}
	}()
}

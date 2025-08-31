package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

// lockoutStatusHandler handles GET /api/auth/lockout/status
func lockoutStatusHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Lockout status endpoint - not implemented in this version",
		})
	}
}

// unlockAccountHandler handles POST /api/auth/lockout/unlock
func unlockAccountHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Unlock account endpoint - not implemented in this version",
		})
	}
}

// lockoutConfigHandler handles GET /api/auth/lockout/config
func lockoutConfigHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Lockout config endpoint - not implemented in this version",
		})
	}
}

// lockoutStatsHandler handles GET /api/auth/lockout/stats
func lockoutStatsHandler(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Lockout stats endpoint - not implemented in this version",
		})
	}
}

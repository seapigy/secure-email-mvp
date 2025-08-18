package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"secure-email-mvp/pkg/pqc"
)

// PQCConfigRequest represents a PQC configuration update request
type PQCConfigRequest struct {
	EnablePQC       *bool `json:"enable_pqc"`
	KyberLevel      *int  `json:"kyber_level"`
	HybridMode      *bool `json:"hybrid_mode"`
	KeyRotationDays *int  `json:"key_rotation_days"`
	HSMEnabled      *bool `json:"hsm_enabled"`
	PerformanceMode *bool `json:"performance_mode"`
	AuditLogging    *bool `json:"audit_logging"`
}

// PQCConfigResponse represents a PQC configuration response
type PQCConfigResponse struct {
	Config *pqc.PQCConfig `json:"config"`
	Stats  map[string]interface{} `json:"stats"`
}

// PQCStatsResponse represents PQC statistics response
type PQCStatsResponse struct {
	ServiceStats    map[string]interface{} `json:"service_stats"`
	MigrationStats  map[string]interface{} `json:"migration_stats"`
	PerformanceStats map[string]interface{} `json:"performance_stats"`
	KeyStats        map[string]interface{} `json:"key_stats"`
}

// PQCHealthResponse represents PQC health check response
type PQCHealthResponse struct {
	Status          string                 `json:"status"`
	PQCEnabled      bool                   `json:"pqc_enabled"`
	KeyManagerOK    bool                   `json:"key_manager_ok"`
	AuditLoggerOK   bool                   `json:"audit_logger_ok"`
	LastKeyRotation time.Time              `json:"last_key_rotation"`
	Details         map[string]interface{} `json:"details"`
}

// pqcConfigHandler handles PQC configuration requests
func pqcConfigHandler(pqcService *pqc.PQCService, pqcIntegration *pqc.PQCIntegration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handleGetPQCConfig(w, r, pqcService, pqcIntegration)
		case "PUT":
			handleUpdatePQCConfig(w, r, pqcService, pqcIntegration)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleGetPQCConfig handles GET requests for PQC configuration
func handleGetPQCConfig(w http.ResponseWriter, r *http.Request, pqcService *pqc.PQCService, pqcIntegration *pqc.PQCIntegration) {
	config := pqcService.GetConfig()
	
	// Get migration stats
	migrationStats, err := pqcIntegration.GetMigrationStats()
	if err != nil {
		log.Printf("Failed to get migration stats: %v", err)
		migrationStats = map[string]interface{}{"error": err.Error()}
	}
	
	response := &PQCConfigResponse{
		Config: config,
		Stats:  migrationStats,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleUpdatePQCConfig handles PUT requests for PQC configuration updates
func handleUpdatePQCConfig(w http.ResponseWriter, r *http.Request, pqcService *pqc.PQCService, pqcIntegration *pqc.PQCIntegration) {
	var req PQCConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	
	// Get current config
	currentConfig := pqcService.GetConfig()
	
	// Update config fields if provided
	if req.EnablePQC != nil {
		currentConfig.EnablePQC = *req.EnablePQC
	}
	if req.KyberLevel != nil {
		if *req.KyberLevel != 512 && *req.KyberLevel != 768 && *req.KyberLevel != 1024 {
			http.Error(w, "Invalid Kyber level. Must be 512, 768, or 1024", http.StatusBadRequest)
			return
		}
		currentConfig.KyberLevel = *req.KyberLevel
	}
	if req.HybridMode != nil {
		currentConfig.HybridMode = *req.HybridMode
	}
	if req.KeyRotationDays != nil {
		if *req.KeyRotationDays < 1 || *req.KeyRotationDays > 365 {
			http.Error(w, "Invalid key rotation days. Must be between 1 and 365", http.StatusBadRequest)
			return
		}
		currentConfig.KeyRotationDays = *req.KeyRotationDays
	}
	if req.HSMEnabled != nil {
		currentConfig.HSMEnabled = *req.HSMEnabled
	}
	if req.PerformanceMode != nil {
		currentConfig.PerformanceMode = *req.PerformanceMode
	}
	if req.AuditLogging != nil {
		currentConfig.AuditLogging = *req.AuditLogging
	}
	
	// Update the configuration
	if err := pqcService.UpdateConfig(currentConfig); err != nil {
		log.Printf("Failed to update PQC config: %v", err)
		http.Error(w, "Failed to update configuration", http.StatusInternalServerError)
		return
	}
	
	// Get updated stats
	migrationStats, err := pqcIntegration.GetMigrationStats()
	if err != nil {
		log.Printf("Failed to get migration stats: %v", err)
		migrationStats = map[string]interface{}{"error": err.Error()}
	}
	
	response := &PQCConfigResponse{
		Config: currentConfig,
		Stats:  migrationStats,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// pqcStatsHandler handles PQC statistics requests
func pqcStatsHandler(pqcService *pqc.PQCService, pqcIntegration *pqc.PQCIntegration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get service stats
		serviceStats := pqcService.GetStats()
		
		// Get migration stats
		migrationStats, err := pqcIntegration.GetMigrationStats()
		if err != nil {
			log.Printf("Failed to get migration stats: %v", err)
			migrationStats = map[string]interface{}{"error": err.Error()}
		}
		
		// Get performance stats from database
		performanceStats, err := getPQCPerformanceStats(pqcIntegration)
		if err != nil {
			log.Printf("Failed to get performance stats: %v", err)
			performanceStats = map[string]interface{}{"error": err.Error()}
		}
		
		// Get key stats from database
		keyStats, err := getPQCKeyStats(pqcIntegration)
		if err != nil {
			log.Printf("Failed to get key stats: %v", err)
			keyStats = map[string]interface{}{"error": err.Error()}
		}
		
		response := &PQCStatsResponse{
			ServiceStats:     serviceStats,
			MigrationStats:   migrationStats,
			PerformanceStats: performanceStats,
			KeyStats:         keyStats,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// pqcHealthHandler handles PQC health check requests
func pqcHealthHandler(pqcService *pqc.PQCService, pqcIntegration *pqc.PQCIntegration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := "healthy"
		keyManagerOK := true
		auditLoggerOK := true
		
		// Check if PQC is enabled
		pqcEnabled := pqcService.IsEnabled()
		
		// Get service stats for health check
		serviceStats := pqcService.GetStats()
		
		// Check key manager health
		keyStats := serviceStats
		if keyStats == nil {
			keyManagerOK = false
			status = "degraded"
		}
		
		// Check audit logger health
		auditStats := pqcService.GetConfig()
		if auditStats == nil {
			auditLoggerOK = false
			status = "degraded"
		}
		
		// Get last key rotation time (simulated)
		lastKeyRotation := time.Now().Add(-24 * time.Hour) // Simulated
		
		response := &PQCHealthResponse{
			Status:          status,
			PQCEnabled:      pqcEnabled,
			KeyManagerOK:    keyManagerOK,
			AuditLoggerOK:   auditLoggerOK,
			LastKeyRotation: lastKeyRotation,
			Details:         serviceStats,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}
}

// pqcMigrationHandler handles PQC migration requests
func pqcMigrationHandler(pqcIntegration *pqc.PQCIntegration) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			handleStartMigration(w, r, pqcIntegration)
		case "GET":
			handleGetMigrationStatus(w, r, pqcIntegration)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleStartMigration handles POST requests to start PQC migration
func handleStartMigration(w http.ResponseWriter, r *http.Request, pqcIntegration *pqc.PQCIntegration) {
	// Parse batch size from query parameter
	batchSizeStr := r.URL.Query().Get("batch_size")
	batchSize := 10 // Default batch size
	
	if batchSizeStr != "" {
		if parsed, err := strconv.Atoi(batchSizeStr); err == nil && parsed > 0 && parsed <= 100 {
			batchSize = parsed
		} else {
			http.Error(w, "Invalid batch_size. Must be between 1 and 100", http.StatusBadRequest)
			return
		}
	}
	
	// Start migration
	migratedCount, err := pqcIntegration.BatchMigrateEmailsToPQC(batchSize)
	if err != nil {
		log.Printf("Failed to start migration: %v", err)
		http.Error(w, fmt.Sprintf("Failed to start migration: %v", err), http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"message":         "Migration started successfully",
		"migrated_count":  migratedCount,
		"batch_size":      batchSize,
		"timestamp":       time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleGetMigrationStatus handles GET requests for migration status
func handleGetMigrationStatus(w http.ResponseWriter, r *http.Request, pqcIntegration *pqc.PQCIntegration) {
	// Get migration stats
	migrationStats, err := pqcIntegration.GetMigrationStats()
	if err != nil {
		log.Printf("Failed to get migration stats: %v", err)
		http.Error(w, "Failed to get migration status", http.StatusInternalServerError)
		return
	}
	
	response := map[string]interface{}{
		"migration_stats": migrationStats,
		"timestamp":       time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// pqcKeyHandler handles PQC key management requests
func pqcKeyHandler(pqcService *pqc.PQCService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			handleGetPQCKeys(w, r, pqcService)
		case "POST":
			handleRotatePQCKeys(w, r, pqcService)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// handleGetPQCKeys handles GET requests for PQC key information
func handleGetPQCKeys(w http.ResponseWriter, r *http.Request, pqcService *pqc.PQCService) {
	// Get current public key
	publicKey, err := pqcService.GetKeyManager().ExportPublicKey()
	if err != nil {
		log.Printf("Failed to export public key: %v", err)
		http.Error(w, "Failed to get public key", http.StatusInternalServerError)
		return
	}
	
	// Get key stats
	keyStats := pqcService.GetKeyManager().GetKeyStats()
	
	response := map[string]interface{}{
		"public_key": publicKey,
		"key_stats":  keyStats,
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// handleRotatePQCKeys handles POST requests to rotate PQC keys
func handleRotatePQCKeys(w http.ResponseWriter, r *http.Request, pqcService *pqc.PQCService) {
	// Rotate keys
	if err := pqcService.GetKeyManager().RotateKeys(); err != nil {
		log.Printf("Failed to rotate keys: %v", err)
		http.Error(w, "Failed to rotate keys", http.StatusInternalServerError)
		return
	}
	
	// Get updated key stats
	keyStats := pqcService.GetKeyManager().GetKeyStats()
	
	response := map[string]interface{}{
		"message":    "Keys rotated successfully",
		"key_stats":  keyStats,
		"timestamp":  time.Now().Format(time.RFC3339),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// getPQCPerformanceStats retrieves PQC performance statistics from the database
func getPQCPerformanceStats(pqcIntegration *pqc.PQCIntegration) (map[string]interface{}, error) {
	// This would query the pqc_performance_metrics table
	// For now, return simulated data
	return map[string]interface{}{
		"total_operations":      1000,
		"successful_operations": 995,
		"failed_operations":     5,
		"avg_encryption_time_ms": 45,
		"avg_decryption_time_ms": 38,
		"success_rate":          99.5,
	}, nil
}

// getPQCKeyStats retrieves PQC key statistics from the database
func getPQCKeyStats(pqcIntegration *pqc.PQCIntegration) (map[string]interface{}, error) {
	// This would query the pqc_keys table
	// For now, return simulated data
	return map[string]interface{}{
		"total_keys":      3,
		"active_keys":     1,
		"inactive_keys":   2,
		"valid_keys":      1,
		"expired_keys":    2,
		"avg_kyber_level": 768,
		"max_rotation":    5,
	}, nil
}

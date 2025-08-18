package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"secure-email-mvp/pkg/email"

	"github.com/gorilla/mux"
)

// RetentionForecastResponse represents the response for retention forecast endpoints
type RetentionForecastResponse struct {
	Success bool                       `json:"success"`
	Data    []*email.RetentionForecast `json:"data,omitempty"`
	Error   string                     `json:"error,omitempty"`
	Meta    *ForecastMeta              `json:"meta,omitempty"`
}

// ForecastMeta represents metadata for forecast responses
type ForecastMeta struct {
	TotalForecasts int     `json:"total_forecasts"`
	AvgConfidence  float64 `json:"avg_confidence"`
	AvgAccuracy    float64 `json:"avg_accuracy"`
	LatestForecast string  `json:"latest_forecast"`
}

// RetentionAnomalyResponse represents the response for retention anomaly endpoints
type RetentionAnomalyResponse struct {
	Success bool                      `json:"success"`
	Data    []*email.RetentionAnomaly `json:"data,omitempty"`
	Error   string                    `json:"error,omitempty"`
	Meta    *AnomalyMeta              `json:"meta,omitempty"`
}

// AnomalyMeta represents metadata for anomaly responses
type AnomalyMeta struct {
	TotalAnomalies    int    `json:"total_anomalies"`
	ActiveAnomalies   int    `json:"active_anomalies"`
	CriticalAnomalies int    `json:"critical_anomalies"`
	LatestAnomaly     string `json:"latest_anomaly"`
}

// AcknowledgeAnomalyRequest represents the request for acknowledging an anomaly
type AcknowledgeAnomalyRequest struct {
	ResolutionNotes string `json:"resolution_notes"`
}

// AcknowledgeAnomalyResponse represents the response for acknowledging an anomaly
type AcknowledgeAnomalyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

// adminRetentionForecastHandler handles GET /api/admin/email/retention-forecast
func (srv *Server) adminRetentionForecastHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	forecastType := r.URL.Query().Get("type")
	forecastKey := r.URL.Query().Get("key")
	limitStr := r.URL.Query().Get("limit")

	// Set default values
	if forecastType == "" {
		forecastType = "global"
	}
	if forecastKey == "" {
		forecastKey = "global"
	}

	limit := 50 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Create forecast service
	forecastService := email.NewRetentionForecastService(srv.db, nil)

	// Get forecasts
	forecasts, err := forecastService.GetForecasts(ctx, forecastType, forecastKey, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve forecasts", err)
		return
	}

	// Calculate metadata
	meta := calculateForecastMeta(forecasts)

	// Prepare response
	response := RetentionForecastResponse{
		Success: true,
		Data:    forecasts,
		Meta:    meta,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// adminRetentionAnomaliesHandler handles GET /api/admin/email/retention-anomalies
func (srv *Server) adminRetentionAnomaliesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	scopeType := r.URL.Query().Get("scope_type")
	scopeKey := r.URL.Query().Get("scope_key")
	severity := r.URL.Query().Get("severity")
	status := r.URL.Query().Get("status")
	limitStr := r.URL.Query().Get("limit")

	limit := 50 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	// Create anomaly detector
	anomalyDetector := email.NewRetentionAnomalyDetector(srv.db, nil)

	// Get anomalies
	anomalies, err := anomalyDetector.GetAnomalies(ctx, scopeType, scopeKey, severity, status, limit)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve anomalies", err)
		return
	}

	// Calculate metadata
	meta := calculateAnomalyMeta(anomalies)

	// Prepare response
	response := RetentionAnomalyResponse{
		Success: true,
		Data:    anomalies,
		Meta:    meta,
	}

	respondWithJSON(w, http.StatusOK, response)
}

// adminAcknowledgeAnomalyHandler handles POST /api/admin/email/retention-anomalies/ack
func (srv *Server) adminAcknowledgeAnomalyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get user from context (set by JWT middleware)
	userID, ok := ctx.Value("user_id").(string)
	if !ok {
		respondWithError(w, http.StatusUnauthorized, "User not authenticated", nil)
		return
	}

	// Parse anomaly ID from URL
	vars := mux.Vars(r)
	anomalyIDStr := vars["id"]
	if anomalyIDStr == "" {
		respondWithError(w, http.StatusBadRequest, "Anomaly ID is required", nil)
		return
	}

	anomalyID, err := strconv.ParseInt(anomalyIDStr, 10, 64)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid anomaly ID", err)
		return
	}

	// Parse request body
	var req AcknowledgeAnomalyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid request body", err)
		return
	}

	// Create anomaly detector
	anomalyDetector := email.NewRetentionAnomalyDetector(srv.db, nil)

	// Acknowledge the anomaly
	if err := anomalyDetector.AcknowledgeAnomaly(ctx, anomalyID, userID, req.ResolutionNotes); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to acknowledge anomaly", err)
		return
	}

	// Prepare response
	response := AcknowledgeAnomalyResponse{
		Success: true,
		Message: "Anomaly acknowledged successfully",
	}

	respondWithJSON(w, http.StatusOK, response)
}

// adminGenerateForecastsHandler handles POST /api/admin/email/retention-forecast/generate
func (srv *Server) adminGenerateForecastsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Create forecast service
	forecastService := email.NewRetentionForecastService(srv.db, nil)

	// Generate forecasts
	if err := forecastService.GenerateForecasts(ctx); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to generate forecasts", err)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"success":   true,
		"message":   "Forecasts generated successfully",
		"timestamp": time.Now().UTC(),
	}

	respondWithJSON(w, http.StatusOK, response)
}

// adminDetectAnomaliesHandler handles POST /api/admin/email/retention-anomalies/detect
func (srv *Server) adminDetectAnomaliesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Create anomaly detector
	anomalyDetector := email.NewRetentionAnomalyDetector(srv.db, nil)

	// Detect anomalies
	if err := anomalyDetector.DetectAnomalies(ctx); err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to detect anomalies", err)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"success":   true,
		"message":   "Anomaly detection completed successfully",
		"timestamp": time.Now().UTC(),
	}

	respondWithJSON(w, http.StatusOK, response)
}

// adminForecastAccuracyHandler handles GET /api/admin/email/retention-forecast/accuracy
func (srv *Server) adminForecastAccuracyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	daysStr := r.URL.Query().Get("days")
	days := 30 // Default to 30 days
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	// Query forecast accuracy data
	query := `
		SELECT 
			COUNT(*) as total_evaluations,
			AVG(overall_accuracy_score) as avg_accuracy,
			AVG(usage_accuracy_percentage) as avg_usage_accuracy,
			AVG(archival_accuracy_percentage) as avg_archival_accuracy,
			AVG(deletion_accuracy_percentage) as avg_deletion_accuracy,
			MAX(evaluated_at) as latest_evaluation
		FROM forecast_accuracy_logs
		WHERE evaluated_at >= datetime('now', '-? days')
	`

	var accuracy struct {
		TotalEvaluations    int       `json:"total_evaluations"`
		AvgAccuracy         float64   `json:"avg_accuracy"`
		AvgUsageAccuracy    float64   `json:"avg_usage_accuracy"`
		AvgArchivalAccuracy float64   `json:"avg_archival_accuracy"`
		AvgDeletionAccuracy float64   `json:"avg_deletion_accuracy"`
		LatestEvaluation    time.Time `json:"latest_evaluation"`
	}

	err := srv.db.QueryRowContext(ctx, query, days).Scan(
		&accuracy.TotalEvaluations,
		&accuracy.AvgAccuracy,
		&accuracy.AvgUsageAccuracy,
		&accuracy.AvgArchivalAccuracy,
		&accuracy.AvgDeletionAccuracy,
		&accuracy.LatestEvaluation,
	)

	if err != nil && err != sql.ErrNoRows {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve forecast accuracy", err)
		return
	}

	// Prepare response
	response := map[string]interface{}{
		"success": true,
		"data":    accuracy,
		"meta": map[string]interface{}{
			"evaluation_period_days": days,
		},
	}

	respondWithJSON(w, http.StatusOK, response)
}

// adminAnomalyStatsHandler handles GET /api/admin/email/retention-anomalies/stats
func (srv *Server) adminAnomalyStatsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	daysStr := r.URL.Query().Get("days")
	days := 30 // Default to 30 days
	if daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 {
			days = d
		}
	}

	// Query anomaly statistics
	query := `
		SELECT 
			anomaly_type,
			severity,
			status,
			COUNT(*) as count,
			AVG(deviation_percentage) as avg_deviation,
			MAX(detected_at) as latest_anomaly
		FROM retention_anomalies
		WHERE detected_at >= datetime('now', '-? days')
		GROUP BY anomaly_type, severity, status
		ORDER BY count DESC
	`

	rows, err := srv.db.QueryContext(ctx, query, days)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Failed to retrieve anomaly statistics", err)
		return
	}
	defer rows.Close()

	var stats []map[string]interface{}
	for rows.Next() {
		var stat struct {
			AnomalyType   string    `json:"anomaly_type"`
			Severity      string    `json:"severity"`
			Status        string    `json:"status"`
			Count         int       `json:"count"`
			AvgDeviation  float64   `json:"avg_deviation"`
			LatestAnomaly time.Time `json:"latest_anomaly"`
		}

		err := rows.Scan(
			&stat.AnomalyType,
			&stat.Severity,
			&stat.Status,
			&stat.Count,
			&stat.AvgDeviation,
			&stat.LatestAnomaly,
		)
		if err != nil {
			continue
		}

		stats = append(stats, map[string]interface{}{
			"anomaly_type":   stat.AnomalyType,
			"severity":       stat.Severity,
			"status":         stat.Status,
			"count":          stat.Count,
			"avg_deviation":  stat.AvgDeviation,
			"latest_anomaly": stat.LatestAnomaly,
		})
	}

	// Calculate summary statistics
	var totalAnomalies, activeAnomalies, criticalAnomalies int
	for _, stat := range stats {
		count := stat["count"].(int)
		severity := stat["severity"].(string)
		status := stat["status"].(string)

		totalAnomalies += count
		if status == "active" {
			activeAnomalies += count
		}
		if severity == "critical" {
			criticalAnomalies += count
		}
	}

	// Prepare response
	response := map[string]interface{}{
		"success": true,
		"data":    stats,
		"meta": map[string]interface{}{
			"total_anomalies":        totalAnomalies,
			"active_anomalies":       activeAnomalies,
			"critical_anomalies":     criticalAnomalies,
			"evaluation_period_days": days,
		},
	}

	respondWithJSON(w, http.StatusOK, response)
}

// Helper functions

// calculateForecastMeta calculates metadata for forecast responses
func calculateForecastMeta(forecasts []*email.RetentionForecast) *ForecastMeta {
	if len(forecasts) == 0 {
		return &ForecastMeta{}
	}

	var totalConfidence, totalAccuracy float64
	var latestForecast time.Time

	for _, forecast := range forecasts {
		totalConfidence += forecast.ConfidenceScore
		totalAccuracy += forecast.AccuracyScore
		if forecast.GeneratedAt.After(latestForecast) {
			latestForecast = forecast.GeneratedAt
		}
	}

	return &ForecastMeta{
		TotalForecasts: len(forecasts),
		AvgConfidence:  totalConfidence / float64(len(forecasts)),
		AvgAccuracy:    totalAccuracy / float64(len(forecasts)),
		LatestForecast: latestForecast.Format(time.RFC3339),
	}
}

// calculateAnomalyMeta calculates metadata for anomaly responses
func calculateAnomalyMeta(anomalies []*email.RetentionAnomaly) *AnomalyMeta {
	if len(anomalies) == 0 {
		return &AnomalyMeta{}
	}

	var activeAnomalies, criticalAnomalies int
	var latestAnomaly time.Time

	for _, anomaly := range anomalies {
		if anomaly.Status == "active" {
			activeAnomalies++
		}
		if anomaly.Severity == "critical" {
			criticalAnomalies++
		}
		if anomaly.DetectedAt.After(latestAnomaly) {
			latestAnomaly = anomaly.DetectedAt
		}
	}

	return &AnomalyMeta{
		TotalAnomalies:    len(anomalies),
		ActiveAnomalies:   activeAnomalies,
		CriticalAnomalies: criticalAnomalies,
		LatestAnomaly:     latestAnomaly.Format(time.RFC3339),
	}
}

// respondWithError sends an error response
func respondWithError(w http.ResponseWriter, statusCode int, message string, err error) {
	errorMsg := message
	if err != nil {
		errorMsg = fmt.Sprintf("%s: %v", message, err)
	}

	response := map[string]interface{}{
		"success": false,
		"error":   errorMsg,
	}

	respondWithJSON(w, statusCode, response)
}

// respondWithJSON sends a JSON response
func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

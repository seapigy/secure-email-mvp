package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"secure-email-mvp/pkg/email"
)

// AdminRetentionInsightsResponse represents the response for retention insights
type AdminRetentionInsightsResponse struct {
	Insights    []*email.RetentionInsight `json:"insights"`
	TotalCount  int                       `json:"total_count"`
	Limit       int                       `json:"limit"`
	Offset      int                       `json:"offset"`
	GeneratedAt time.Time                 `json:"generated_at"`
}

// AdminRetentionTrendsResponse represents the response for trend analysis
type AdminRetentionTrendsResponse struct {
	TrendAnalysis map[string]interface{} `json:"trend_analysis"`
	DateRange     map[string]string      `json:"date_range"`
	GeneratedAt   time.Time              `json:"generated_at"`
}

// AdminRetentionRecommendationsResponse represents the response for recommendations
type AdminRetentionRecommendationsResponse struct {
	Recommendations []*email.RetentionRecommendation `json:"recommendations"`
	TotalCount      int                              `json:"total_count"`
	Limit           int                              `json:"limit"`
	Offset          int                              `json:"offset"`
	GeneratedAt     time.Time                        `json:"generated_at"`
}

// ApplyRecommendationRequest represents the request for applying a recommendation
type ApplyRecommendationRequest struct {
	RecommendationID int64 `json:"recommendation_id"`
	Preview          bool  `json:"preview"` // If true, only preview the changes
}

// ApplyRecommendationResponse represents the response for applying a recommendation
type ApplyRecommendationResponse struct {
	Success          bool                   `json:"success"`
	Message          string                 `json:"message"`
	RecommendationID int64                  `json:"recommendation_id"`
	PreviewMode      bool                   `json:"preview_mode"`
	Result           map[string]interface{} `json:"result"`
	AppliedAt        time.Time              `json:"applied_at"`
}

// adminRetentionInsightsHandler handles GET /api/admin/email/retention-insights
func (srv *Server) adminRetentionInsightsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("adminRetentionInsightsHandler started - Admin Retention Insights (Micro-Iteration 4.27)")

	// Check authentication
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	log.Printf("Admin retention insights requested by user: %s", userID)

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	insightType := r.URL.Query().Get("insight_type")
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")

	// Set defaults and validate
	limit := 50 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Build filters
	filters := make(map[string]string)
	if insightType != "" {
		filters["insight_type"] = insightType
	}
	if startDateStr != "" {
		filters["start_date"] = startDateStr
	}
	if endDateStr != "" {
		filters["end_date"] = endDateStr
	}

	// Initialize insights service
	insightsService := email.NewRetentionInsightsService(srv.db)

	// Get insights
	insights, err := insightsService.GetInsights(r.Context(), filters, limit, offset)
	if err != nil {
		log.Printf("Failed to get retention insights: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve insights"}`))
		return
	}

	// Get total count (without limit/offset)
	totalInsights, err := insightsService.GetInsights(r.Context(), filters, 0, 0)
	if err != nil {
		log.Printf("Failed to get total count: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve total count"}`))
		return
	}

	response := AdminRetentionInsightsResponse{
		Insights:    insights,
		TotalCount:  len(totalInsights),
		Limit:       limit,
		Offset:      offset,
		GeneratedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// adminRetentionTrendsHandler handles GET /api/admin/email/retention-insights/trends
func (srv *Server) adminRetentionTrendsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("adminRetentionTrendsHandler started - Admin Retention Trends (Micro-Iteration 4.27)")

	// Check authentication
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	log.Printf("Admin retention trends requested by user: %s", userID)

	// Parse query parameters
	startDateStr := r.URL.Query().Get("start_date")
	endDateStr := r.URL.Query().Get("end_date")
	exportCSV := r.URL.Query().Get("export_csv") == "true"

	// Set default date range (last 30 days)
	now := time.Now()
	endDate := now
	startDate := now.AddDate(0, 0, -30)

	// Parse custom date range if provided
	if startDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", startDateStr); err == nil {
			startDate = parsed
		}
	}

	if endDateStr != "" {
		if parsed, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endDate = parsed
		}
	}

	// Initialize insights service
	insightsService := email.NewRetentionInsightsService(srv.db)

	// Get trend analysis
	trendAnalysis, err := insightsService.GetTrendAnalysis(r.Context(), startDate, endDate)
	if err != nil {
		log.Printf("Failed to get trend analysis: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve trend analysis"}`))
		return
	}

	// Handle CSV export
	if exportCSV {
		w.Header().Set("Content-Type", "text/csv")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=retention_trends_%s_to_%s.csv", 
			startDate.Format("2006-01-02"), endDate.Format("2006-01-02")))
		
		// Generate CSV content
		csvContent := generateTrendsCSV(trendAnalysis)
		w.Write([]byte(csvContent))
		return
	}

	response := AdminRetentionTrendsResponse{
		TrendAnalysis: trendAnalysis,
		DateRange: map[string]string{
			"start_date": startDate.Format("2006-01-02"),
			"end_date":   endDate.Format("2006-01-02"),
		},
		GeneratedAt: time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// adminRetentionRecommendationsHandler handles GET /api/admin/email/retention-recommendations
func (srv *Server) adminRetentionRecommendationsHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("adminRetentionRecommendationsHandler started - Admin Retention Recommendations (Micro-Iteration 4.27)")

	// Check authentication
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	log.Printf("Admin retention recommendations requested by user: %s", userID)

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	recommendationType := r.URL.Query().Get("recommendation_type")
	priority := r.URL.Query().Get("priority")
	status := r.URL.Query().Get("status")
	userIDFilter := r.URL.Query().Get("user_id")
	domain := r.URL.Query().Get("domain")

	// Set defaults and validate
	limit := 50 // Default limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 1000 {
			limit = l
		}
	}

	offset := 0
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Build filters
	filters := make(map[string]string)
	if recommendationType != "" {
		filters["recommendation_type"] = recommendationType
	}
	if priority != "" {
		filters["priority"] = priority
	}
	if status != "" {
		filters["status"] = status
	}
	if userIDFilter != "" {
		filters["user_id"] = userIDFilter
	}
	if domain != "" {
		filters["domain"] = domain
	}

	// Initialize advisor service
	advisorService := email.NewRetentionAdvisorService(srv.db)

	// Get recommendations
	recommendations, err := advisorService.GetRecommendations(r.Context(), filters, limit, offset)
	if err != nil {
		log.Printf("Failed to get retention recommendations: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve recommendations"}`))
		return
	}

	// Get total count (without limit/offset)
	totalRecommendations, err := advisorService.GetRecommendations(r.Context(), filters, 0, 0)
	if err != nil {
		log.Printf("Failed to get total count: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"Failed to retrieve total count"}`))
		return
	}

	response := AdminRetentionRecommendationsResponse{
		Recommendations: recommendations,
		TotalCount:      len(totalRecommendations),
		Limit:           limit,
		Offset:          offset,
		GeneratedAt:     time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// adminApplyRecommendationHandler handles POST /api/admin/email/retention-recommendations/apply
func (srv *Server) adminApplyRecommendationHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("adminApplyRecommendationHandler started - Admin Apply Recommendation (Micro-Iteration 4.27)")

	// Check authentication
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		log.Printf("User ID not found in context")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Authentication required"}`))
		return
	}

	log.Printf("Admin apply recommendation requested by user: %s", userID)

	// Parse request body
	var req ApplyRecommendationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Failed to decode request body: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid request body"}`))
		return
	}

	// Validate request
	if req.RecommendationID <= 0 {
		log.Printf("Invalid recommendation ID: %d", req.RecommendationID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error":"Invalid recommendation ID"}`))
		return
	}

	// Initialize advisor service
	advisorService := email.NewRetentionAdvisorService(srv.db)

	// Apply recommendation
	result, err := advisorService.ApplyRecommendation(r.Context(), req.RecommendationID, userID, req.Preview)
	if err != nil {
		log.Printf("Failed to apply recommendation: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf(`{"error":"Failed to apply recommendation: %v"}`, err)))
		return
	}

	response := ApplyRecommendationResponse{
		Success:          true,
		Message:          "Recommendation applied successfully",
		RecommendationID: req.RecommendationID,
		PreviewMode:      req.Preview,
		Result:           result,
		AppliedAt:        time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// generateTrendsCSV generates CSV content for trend analysis export
func generateTrendsCSV(trendAnalysis map[string]interface{}) string {
	csv := "Date,Policy Effectiveness,Compression Ratio,Storage Savings (bytes),Cost Savings (USD),Override Frequency\n"
	
	if dailyTrends, ok := trendAnalysis["daily_trends"].([]map[string]interface{}); ok {
		for _, trend := range dailyTrends {
			date := trend["date"].(string)
			effectiveness := fmt.Sprintf("%.3f", trend["policy_effectiveness"].(float64))
			compression := fmt.Sprintf("%.3f", trend["compression_ratio"].(float64))
			storageSavings := fmt.Sprintf("%d", int64(trend["storage_savings_bytes"].(float64)))
			costSavings := fmt.Sprintf("%.2f", trend["cost_savings_usd"].(float64))
			overrides := fmt.Sprintf("%d", int(trend["override_frequency"].(float64)))
			
			csv += fmt.Sprintf("%s,%s,%s,%s,%s,%s\n", 
				date, effectiveness, compression, storageSavings, costSavings, overrides)
		}
	}
	
	return csv
}

package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"secure-email-mvp/pkg/securelinks"

	"github.com/gorilla/mux"
)

// =============================================================================
// SECURE LINKS API HANDLERS
// =============================================================================

// createSecureLinkHandler handles POST /api/secure-links
func (srv *Server) createSecureLinkHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by JWT middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse request body
	var req securelinks.CreateSecureLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Create secure link
	response, err := srv.secureLinksService.CreateSecureLink(r.Context(), req, userID)
	if err != nil {
		// Log error for debugging
		log.Printf("Error creating secure link: %v", err)

		// Return appropriate error response
		if strings.Contains(err.Error(), "email not found") {
			http.Error(w, "Email not found or access denied", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "invalid request") {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// accessSecureLinkHandler handles POST /api/secure-links/{linkID}/access
func (srv *Server) accessSecureLinkHandler(w http.ResponseWriter, r *http.Request) {
	// Extract link ID from URL
	vars := mux.Vars(r)
	linkID := vars["linkID"]
	if linkID == "" {
		http.Error(w, "Link ID is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var req securelinks.AccessSecureLinkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Set link ID from URL
	req.LinkID = linkID

	// Get client IP address
	req.IPAddress = getClientIPForSecureLinks(r)
	req.UserAgent = r.UserAgent()

	// Access secure link
	response, err := srv.secureLinksService.AccessSecureLink(r.Context(), req)
	if err != nil {
		// Log error for debugging
		log.Printf("Error accessing secure link %s: %v", linkID, err)

		// Return appropriate error response
		if strings.Contains(err.Error(), "link not found") {
			http.Error(w, "Link not found or access denied", http.StatusNotFound)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	if !response.Success {
		w.WriteHeader(http.StatusForbidden)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	json.NewEncoder(w).Encode(response)
}

// revokeSecureLinkHandler handles DELETE /api/secure-links/{linkID}
func (srv *Server) revokeSecureLinkHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by JWT middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract link ID from URL
	vars := mux.Vars(r)
	linkID := vars["linkID"]
	if linkID == "" {
		http.Error(w, "Link ID is required", http.StatusBadRequest)
		return
	}

	// Revoke secure link
	err := srv.secureLinksService.RevokeSecureLink(r.Context(), linkID, userID)
	if err != nil {
		// Log error for debugging
		log.Printf("Error revoking secure link %s: %v", linkID, err)

		// Return appropriate error response
		if strings.Contains(err.Error(), "link not found") {
			http.Error(w, "Link not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "unauthorized") {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Secure link revoked successfully",
		"link_id": linkID,
	})
}

// getSecureLinkInfoHandler handles GET /api/secure-links/{linkID}
func (srv *Server) getSecureLinkInfoHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by JWT middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Extract link ID from URL
	vars := mux.Vars(r)
	linkID := vars["linkID"]
	if linkID == "" {
		http.Error(w, "Link ID is required", http.StatusBadRequest)
		return
	}

	// Get secure link info
	link, err := srv.secureLinksService.GetSecureLinkInfo(r.Context(), linkID, userID)
	if err != nil {
		// Log error for debugging
		log.Printf("Error getting secure link info %s: %v", linkID, err)

		// Return appropriate error response
		if strings.Contains(err.Error(), "link not found") {
			http.Error(w, "Link not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "unauthorized") {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}

		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(link)
}

// listSecureLinksHandler handles GET /api/secure-links
func (srv *Server) listSecureLinksHandler(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by JWT middleware)
	userID, ok := r.Context().Value("user_id").(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	// Set default values
	limit := 50
	offset := 0

	// Parse limit
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	// Parse offset
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// List secure links
	links, err := srv.secureLinksService.ListSecureLinks(r.Context(), userID, limit, offset)
	if err != nil {
		// Log error for debugging
		log.Printf("Error listing secure links for user %s: %v", userID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"links":  links,
		"limit":  limit,
		"offset": offset,
		"count":  len(links),
	})
}

// getSecureLinkTemplateHandler handles GET /api/secure-links/templates
func (srv *Server) getSecureLinkTemplateHandler(w http.ResponseWriter, r *http.Request) {
	// Get default template
	template := &securelinks.SecureLinkTemplate{
		ID:              "default-external-template",
		TemplateName:    "Default External Message",
		TemplateContent: "You have received a secure email from {{sender_name}}. Click the link below to view it securely. This link will expire and includes security features to protect your privacy.",
		IsDefault:       true,
		IsActive:        true,
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(template)
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// getClientIPForSecureLinks extracts the client IP address from the request
func getClientIPForSecureLinks(r *http.Request) string {
	// Check for X-Forwarded-For header (for proxy/load balancer scenarios)
	if forwardedFor := r.Header.Get("X-Forwarded-For"); forwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		if commaIndex := strings.Index(forwardedFor, ","); commaIndex != -1 {
			return strings.TrimSpace(forwardedFor[:commaIndex])
		}
		return strings.TrimSpace(forwardedFor)
	}

	// Check for X-Real-IP header
	if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	if r.RemoteAddr != "" {
		// Remove port if present
		if colonIndex := strings.LastIndex(r.RemoteAddr, ":"); colonIndex != -1 {
			return r.RemoteAddr[:colonIndex]
		}
		return r.RemoteAddr
	}

	// Default fallback
	return "unknown"
}

// =============================================================================
// MIDDLEWARE FOR SECURE LINKS
// =============================================================================

// secureLinkRateLimitMiddleware applies rate limiting to secure link endpoints
func (srv *Server) secureLinkRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: Implement proper rate limiting in Phase 2
		// For now, allow all requests

		next.ServeHTTP(w, r)
	})
}

// secureLinkSecurityHeadersMiddleware adds security headers for secure link endpoints
func (srv *Server) secureLinkSecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Add security headers
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		// Add CSP header for secure link pages
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self';")

		next.ServeHTTP(w, r)
	})
}

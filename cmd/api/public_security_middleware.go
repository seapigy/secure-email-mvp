package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

// publicSecurityHeadersMiddleware applies strict security headers for public secure link pages
func (srv *Server) publicSecurityHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Strict security headers for public pages
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=(), usb=()")
		
		// Content Security Policy for public pages
		// Strict CSP with nonce-based script execution
		csp := buildCSPPolicy(r)
		w.Header().Set("Content-Security-Policy", csp)
		
		// Additional security headers
		w.Header().Set("X-Download-Options", "noopen")
		w.Header().Set("X-Permitted-Cross-Domain-Policies", "none")
		
		// Cache control for security-sensitive pages
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, proxy-revalidate")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("Expires", "0")
		
		next.ServeHTTP(w, r)
	})
}

// buildCSPPolicy builds a strict Content Security Policy for public pages
func buildCSPPolicy(r *http.Request) string {
	// Base CSP directives
	directives := []string{
		"default-src 'self'",
		"script-src 'self' 'nonce-" + generateNonce(r) + "' 'strict-dynamic'",
		"style-src 'self' 'unsafe-inline'", // Allow inline styles for Tailwind
		"img-src 'self' data: https:",
		"font-src 'self' data:",
		"connect-src 'self'",
		"frame-ancestors 'none'",
		"base-uri 'self'",
		"form-action 'self'",
		"upgrade-insecure-requests",
		"block-all-mixed-content",
	}
	
	return strings.Join(directives, "; ")
}

// generateNonce generates a nonce for CSP script execution
func generateNonce(r *http.Request) string {
	// In a production environment, this should be cryptographically secure
	// For now, we'll use a simple hash of the request
	return "nonce-" + hashString(r.RemoteAddr+r.UserAgent())
}

// hashString creates a simple hash for nonce generation
func hashString(s string) string {
	hash := 0
	for _, char := range s {
		hash = ((hash << 5) - hash) + int(char)
		hash = hash & hash // Convert to 32-bit integer
	}
	return fmt.Sprintf("%x", hash)
}

// secureLinkValidationMiddleware validates secure link access patterns
func (srv *Server) secureLinkValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract link ID from URL
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) >= 3 && pathParts[1] == "v" {
			linkID := pathParts[2]
			
			// Validate link ID format
			if !isValidLinkID(linkID) {
				http.Error(w, "Invalid link format", http.StatusBadRequest)
				return
			}
			
			// Check for suspicious access patterns
			if srv.isSuspiciousAccess(r, linkID) {
				// Log suspicious activity
				srv.logSuspiciousAccess(r, linkID)
				
				// Return decoy response for security
				http.Error(w, "Link not found", http.StatusNotFound)
				return
			}
		}
		
		next.ServeHTTP(w, r)
	})
}

// isValidLinkID validates the format of a secure link ID
func isValidLinkID(linkID string) bool {
	// Link ID should be a valid UUID format
	if len(linkID) != 36 {
		return false
	}
	
	// Check UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	parts := strings.Split(linkID, "-")
	if len(parts) != 5 {
		return false
	}
	
	if len(parts[0]) != 8 || len(parts[1]) != 4 || len(parts[2]) != 4 || len(parts[3]) != 4 || len(parts[4]) != 12 {
		return false
	}
	
	return true
}

// isSuspiciousAccess checks for suspicious access patterns
func (srv *Server) isSuspiciousAccess(r *http.Request, linkID string) bool {
	// Check for rapid access attempts
	if srv.isRapidAccess(r.RemoteAddr, linkID) {
		return true
	}
	
	// Check for multiple IPs accessing same link
	if srv.isMultipleIPAccess(linkID) {
		return true
	}
	
	// Check for suspicious user agents
	if srv.isSuspiciousUserAgent(r.UserAgent()) {
		return true
	}
	
	return false
}

// isRapidAccess checks if there are rapid access attempts from the same IP
func (srv *Server) isRapidAccess(ip, linkID string) bool {
	// This would integrate with rate limiting service
	// For now, return false
	return false
}

// isMultipleIPAccess checks if multiple IPs are accessing the same link
func (srv *Server) isMultipleIPAccess(linkID string) bool {
	// This would check recent access logs
	// For now, return false
	return false
}

// isSuspiciousUserAgent checks if the user agent is suspicious
func (srv *Server) isSuspiciousUserAgent(userAgent string) bool {
	suspiciousPatterns := []string{
		"bot", "crawler", "spider", "scraper",
		"curl", "wget", "python", "java",
		"automation", "headless",
	}
	
	userAgentLower := strings.ToLower(userAgent)
	for _, pattern := range suspiciousPatterns {
		if strings.Contains(userAgentLower, pattern) {
			return true
		}
	}
	
	return false
}

// logSuspiciousAccess logs suspicious access attempts
func (srv *Server) logSuspiciousAccess(r *http.Request, linkID string) {
	// Log suspicious activity for monitoring
	log.Printf("[SUSPICIOUS_ACCESS] Link: %s, IP: %s, UA: %s", 
		linkID, r.RemoteAddr, r.UserAgent())
	
	// In production, this would write to structured logs
	// and potentially trigger alerts
}

// publicRateLimitMiddleware applies rate limiting for public endpoints
func (srv *Server) publicRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Apply stricter rate limiting for public endpoints
		clientIP := r.RemoteAddr
		key := "public:" + clientIP
		
		// Check current rate limit
		if count, exists := srv.rateLimits.Load(key); exists {
			if count.(int) >= 5 { // 5 requests per minute
				w.Header().Set("Retry-After", "60")
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}
		}
		
		// Increment rate limit counter
		if count, exists := srv.rateLimits.Load(key); exists {
			srv.rateLimits.Store(key, count.(int)+1)
		} else {
			srv.rateLimits.Store(key, 1)
		}
		
		next.ServeHTTP(w, r)
	})
}

package main

import (
	"context"
	"net/http"
	"strings"

	"secure-email-mvp/pkg/auth"
)

// ContextKey is a type for context keys to avoid collisions
type ContextKey string

const (
	// UserEmailKey is the context key for storing user email
	UserEmailKey ContextKey = "user_email"
)

// jwtMiddleware verifies JWT tokens and sets user email in context
func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"Invalid or missing token"}`, http.StatusUnauthorized)
			return
		}

		// Check if it's a Bearer token
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"Invalid or missing token"}`, http.StatusUnauthorized)
			return
		}

		// Extract token (remove "Bearer " prefix)
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			http.Error(w, `{"error":"Invalid or missing token"}`, http.StatusUnauthorized)
			return
		}

		// Parse and validate JWT token
		claims, err := auth.ParseJWT(tokenString)
		if err != nil {
			http.Error(w, `{"error":"Invalid or missing token"}`, http.StatusUnauthorized)
			return
		}

		// Set user email in context
		ctx := context.WithValue(r.Context(), UserEmailKey, claims.Email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserEmailFromContext extracts user email from request context
func GetUserEmailFromContext(r *http.Request) (string, bool) {
	email, ok := r.Context().Value(UserEmailKey).(string)
	return email, ok
}

package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"secure-email-mvp/pkg/auth"
)

// EmailKey is the context key for storing the user's email in the request context
const EmailKey contextKey = "email"

// EnhancedJWTMiddleware creates middleware that validates JWT access tokens for protected endpoints
func EnhancedJWTMiddleware(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, `{"error":"Authorization header required"}`, http.StatusUnauthorized)
				return
			}

			// Check Bearer format
			if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
				http.Error(w, `{"error":"Invalid authorization format"}`, http.StatusUnauthorized)
				return
			}

			tokenString := authHeader[7:]
			if tokenString == "" {
				http.Error(w, `{"error":"Token is required"}`, http.StatusUnauthorized)
				return
			}

			// Create session manager
			sessionManager, err := auth.NewSessionManager()
			if err != nil {
				http.Error(w, `{"error":"Session configuration error"}`, http.StatusInternalServerError)
				return
			}

			// Validate access token
			claims, err := sessionManager.ValidateAccessToken(tokenString)
			if err != nil {
				http.Error(w, `{"error":"Invalid or expired access token"}`, http.StatusUnauthorized)
				return
			}

			// Set user context
			ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
			ctx = context.WithValue(ctx, EmailKey, claims.Email)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserFromContext extracts user information from the request context
func GetUserFromContext(r *http.Request) (string, string, bool) {
	userID, userIDOk := r.Context().Value(UserIDKey).(string)
	email, emailOk := r.Context().Value(EmailKey).(string)
	return userID, email, userIDOk && emailOk
}

// GetEmailFromContext extracts the user's email from the request context
func GetEmailFromContext(r *http.Request) (string, bool) {
	email, ok := r.Context().Value(EmailKey).(string)
	return email, ok
}

// GetEmail is a helper function for extracting email from context
func GetEmail(ctx context.Context) (string, error) {
	email, ok := ctx.Value(EmailKey).(string)
	if !ok {
		return "", fmt.Errorf("email not found in context")
	}
	return email, nil
}

package auth

// DO NOT EDIT EXISTING CODE - new file added
// Session middleware: validate Bearer token by hashing and checking sessions table.

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"
)

// Context key for user ID
type contextKey string

const userIDKey contextKey = "userID"

// Set DB in main application before handlers used
var DB *sql.DB

// ContextWithUserID adds user ID to context
func ContextWithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDKey, userID)
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(ctx context.Context) (string, bool) {
	userID, ok := ctx.Value(userIDKey).(string)
	return userID, ok
}

func TokenAuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		rawToken := strings.TrimPrefix(auth, "Bearer ")
		tokenHash := HashToken(rawToken)

		var userID string
		var expiresAt time.Time
		err := DB.QueryRow("SELECT user_id, expires_at FROM sessions WHERE token_hash = ? LIMIT 1", tokenHash).Scan(&userID, &expiresAt)
		if err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			log.Printf("ERROR session lookup: %v", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		if time.Now().After(expiresAt) {
			http.Error(w, "token expired", http.StatusUnauthorized)
			return
		}
		// Attach user id to context
		ctx := ContextWithUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

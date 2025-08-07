package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}

// AccessTokenClaims represents the claims for JWT access tokens
type AccessTokenClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.StandardClaims
}

// RefreshTokenClaims represents the claims for JWT refresh tokens
type RefreshTokenClaims struct {
	UserID  string `json:"user_id"`
	TokenID string `json:"token_id"`
	jwt.StandardClaims
}

// SessionManager handles JWT token generation and validation
type SessionManager struct {
	accessTokenSecret  string
	refreshTokenSecret string
	accessTokenExpiry  time.Duration
	refreshTokenExpiry time.Duration
}

// NewSessionManager creates a new session manager with configured secrets and expiry times
func NewSessionManager() (*SessionManager, error) {
	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	if accessSecret == "" {
		accessSecret = os.Getenv("JWT_SECRET") // Fallback to existing secret
	}
	if accessSecret == "" {
		return nil, fmt.Errorf("JWT_ACCESS_SECRET or JWT_SECRET not configured")
	}

	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if refreshSecret == "" {
		refreshSecret = accessSecret // Use same secret if not configured separately
	}

	return &SessionManager{
		accessTokenSecret:  accessSecret,
		refreshTokenSecret: refreshSecret,
		accessTokenExpiry:  15 * time.Minute,   // 15 minutes
		refreshTokenExpiry: 7 * 24 * time.Hour, // 7 days
	}, nil
}

// GenerateTokenPair creates both access and refresh tokens for a user
func (sm *SessionManager) GenerateTokenPair(userID, email string, db *sql.DB) (*TokenPair, error) {
	// Generate access token
	accessToken, err := sm.GenerateAccessToken(userID, email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Generate refresh token
	refreshToken, err := sm.generateRefreshToken(userID, db)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(sm.accessTokenExpiry.Seconds()),
	}, nil
}

// GenerateAccessToken creates a short-lived JWT access token
func (sm *SessionManager) GenerateAccessToken(userID, email string) (string, error) {
	now := time.Now()
	claims := &AccessTokenClaims{
		UserID: userID,
		Email:  email,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: now.Add(sm.accessTokenExpiry).Unix(),
			IssuedAt:  now.Unix(),
			Issuer:    "secure-email-mvp",
			Subject:   userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(sm.accessTokenSecret))
}

// generateRefreshToken creates a long-lived refresh token and stores it in the database
func (sm *SessionManager) generateRefreshToken(userID string, db *sql.DB) (string, error) {
	// Generate a random refresh token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate random token: %w", err)
	}
	refreshToken := base64.URLEncoding.EncodeToString(tokenBytes)

	// Hash the token for secure storage
	tokenHash := sha256.Sum256([]byte(refreshToken))
	tokenHashB64 := base64.StdEncoding.EncodeToString(tokenHash[:])

	// Generate token ID
	tokenID := uuid.New().String()

	// Create JWT refresh token with token ID
	now := time.Now()
	claims := &RefreshTokenClaims{
		UserID:  userID,
		TokenID: tokenID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: now.Add(sm.refreshTokenExpiry).Unix(),
			IssuedAt:  now.Unix(),
			Issuer:    "secure-email-mvp",
			Subject:   userID,
		},
	}

	jwtRefreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	jwtTokenString, err := jwtRefreshToken.SignedString([]byte(sm.refreshTokenSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign refresh token: %w", err)
	}

	// Store refresh token hash in database
	_, err = db.Exec(`
		INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, created_at, last_used_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		tokenID, userID, tokenHashB64, now.Add(sm.refreshTokenExpiry), now, now,
	)
	if err != nil {
		return "", fmt.Errorf("failed to store refresh token: %w", err)
	}

	return jwtTokenString, nil
}

// ValidateAccessToken validates and parses an access token
func (sm *SessionManager) ValidateAccessToken(tokenString string) (*AccessTokenClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(sm.accessTokenSecret), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse access token: %w", err)
	}

	if claims, ok := token.Claims.(*AccessTokenClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, fmt.Errorf("invalid access token")
}

// ValidateRefreshToken validates a refresh token and returns the user ID
func (sm *SessionManager) ValidateRefreshToken(tokenString string, db *sql.DB) (string, error) {
	// Parse the JWT refresh token
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(sm.refreshTokenSecret), nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse refresh token: %w", err)
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid refresh token")
	}

	// Check if token is revoked in database
	var isRevoked bool
	err = db.QueryRow("SELECT is_revoked FROM refresh_tokens WHERE id = ?", claims.TokenID).Scan(&isRevoked)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", fmt.Errorf("refresh token not found")
		}
		return "", fmt.Errorf("database error: %w", err)
	}

	if isRevoked {
		return "", fmt.Errorf("refresh token has been revoked")
	}

	// Update last used timestamp
	_, err = db.Exec("UPDATE refresh_tokens SET last_used_at = ? WHERE id = ?", time.Now(), claims.TokenID)
	if err != nil {
		// Log error but don't fail the validation
		fmt.Printf("Warning: failed to update refresh token last_used_at: %v\n", err)
	}

	return claims.UserID, nil
}

// RevokeRefreshToken marks a refresh token as revoked
func (sm *SessionManager) RevokeRefreshToken(tokenString string, db *sql.DB) error {
	// Parse the JWT to get the token ID
	token, err := jwt.ParseWithClaims(tokenString, &RefreshTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(sm.refreshTokenSecret), nil
	})

	if err != nil {
		return fmt.Errorf("failed to parse refresh token: %w", err)
	}

	claims, ok := token.Claims.(*RefreshTokenClaims)
	if !ok || !token.Valid {
		return fmt.Errorf("invalid refresh token")
	}

	// Mark token as revoked in database
	_, err = db.Exec("UPDATE refresh_tokens SET is_revoked = TRUE WHERE id = ?", claims.TokenID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}

// RevokeAllUserTokens revokes all refresh tokens for a specific user
func (sm *SessionManager) RevokeAllUserTokens(userID string, db *sql.DB) error {
	_, err := db.Exec("UPDATE refresh_tokens SET is_revoked = TRUE WHERE user_id = ?", userID)
	if err != nil {
		return fmt.Errorf("failed to revoke user tokens: %w", err)
	}
	return nil
}

// GetAccessTokenExpiry returns the access token expiry duration
func (sm *SessionManager) GetAccessTokenExpiry() time.Duration {
	return sm.accessTokenExpiry
}

// CleanupExpiredTokens removes expired tokens from the database
func (sm *SessionManager) CleanupExpiredTokens(db *sql.DB) error {
	_, err := db.Exec("DELETE FROM refresh_tokens WHERE expires_at < ?", time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", err)
	}
	return nil
}

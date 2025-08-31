package admin

import (
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	"secure-email-mvp/pkg/models"
)

// Service handles admin authentication and management
type Service struct {
	db         *sql.DB
	repository Repository
}

// NewService creates a new admin service
func NewService(db *sql.DB, repo Repository) *Service {
	return &Service{
		db:         db,
		repository: repo,
	}
}

// LoginAdmin authenticates an admin user and returns a JWT token
func (s *Service) LoginAdmin(email, password string) (string, *models.AdminUser, error) {
	// Get admin user by email
	admin, err := s.repository.GetAdminByEmail(email)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get admin: %w", err)
	}
	
	if admin == nil {
		return "", nil, fmt.Errorf("admin not found")
	}
	
	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password)); err != nil {
		return "", nil, fmt.Errorf("invalid password")
	}
	
	// Verify role is admin
	if admin.Role != "admin" {
		return "", nil, fmt.Errorf("invalid role")
	}
	
	// Generate JWT token
	token, err := s.generateJWT(admin)
	if err != nil {
		return "", nil, fmt.Errorf("failed to generate token: %w", err)
	}
	
	return token, admin, nil
}

// generateJWT creates a JWT token for the admin user
func (s *Service) generateJWT(admin *models.AdminUser) (string, error) {
	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", fmt.Errorf("JWT_SECRET not configured")
	}
	
	// Create JWT claims
	now := time.Now()
	exp := now.Add(24 * time.Hour) // 24 hour expiry
	
	claims := jwt.MapClaims{
		"sub":   fmt.Sprintf("admin:%d", admin.ID),
		"email": admin.Email,
		"role":  admin.Role,
		"exp":   exp.Unix(),
		"iat":   now.Unix(),
		"nbf":   now.Unix(),
	}
	
	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// Sign the token
	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", fmt.Errorf("failed to sign JWT: %w", err)
	}
	
	return tokenString, nil
}

// CreateAdmin creates a new admin user with hashed password
func (s *Service) CreateAdmin(email, password string) (*models.AdminUser, error) {
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Create admin user
	admin := &models.AdminUser{
		Email:     email,
		Password:  string(hashedPassword),
		Role:      "admin",
		CreatedAt: time.Now(),
	}
	
	// Save to database
	if err := s.repository.CreateAdmin(admin); err != nil {
		return nil, fmt.Errorf("failed to create admin: %w", err)
	}
	
	return admin, nil
}

// generateRandomBytes generates random bytes for token generation
func generateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

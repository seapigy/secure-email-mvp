package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"secure-email-mvp/pkg/auth"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

// main is the entry point for the Secure Email API server. It loads environment variables, connects to the database,
// applies schema migrations, sets up middleware, and registers all API endpoints including authentication and fallback flows.
func main() {
	// Load .env with detailed logging
	log.Printf("Starting Secure Email API server...")
	log.Printf("Current working directory: %s", getCurrentDir())

	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
		log.Printf("Attempting to load .env from /home/opc/secure-email-mvp/.env")
		if err := godotenv.Load("/home/opc/secure-email-mvp/.env"); err != nil {
			log.Fatal("Error loading .env from /home/opc/secure-email-mvp/.env:", err)
		}
		log.Printf("Successfully loaded .env from /home/opc/secure-email-mvp/.env")
	} else {
		log.Printf("Successfully loaded .env from current directory")
	}

	// Verify JWT_SECRET is loaded
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Printf("Warning: JWT_SECRET not found in environment, using default")
	} else {
		log.Printf("JWT_SECRET loaded successfully")
	}

	// Connect to SQLite with detailed logging
	dbPath := os.Getenv("SQLITE_DB")
	if dbPath == "" {
		dbPath = "/var/db/secure-email.db"
	}
	log.Printf("Attempting to connect to database at: %s", dbPath)

	// Ensure database directory exists
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Printf("Warning: Could not create database directory %s: %v", dbDir, err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	// Test database connection
	log.Printf("Testing database connection...")
	if err := db.Ping(); err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	log.Printf("Database connection successful")

	// Test R2 connection
	if err := testR2Connection(); err != nil {
		log.Printf("Warning: R2 connection test failed: %v", err)
	} else {
		log.Printf("R2 connection test passed")
	}

	// Initialize server
	srv := &Server{db: db, rateLimits: &sync.Map{}}

	// Apply schema with detailed logging
	log.Printf("Loading database schema...")
	schemaPath := "schema/users_simple.sql"
	log.Printf("Attempting to read schema from: %s", schemaPath)

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		log.Printf("Error reading schema from %s: %v", schemaPath, err)
		log.Printf("Attempting to read schema from absolute path...")
		absPath := filepath.Join(getCurrentDir(), "schema", "users_simple.sql")
		log.Printf("Trying absolute path: %s", absPath)
		schema, err = os.ReadFile(absPath)
		if err != nil {
			log.Fatal("Error reading schema from absolute path:", err)
		}
	}

	log.Printf("Schema file loaded successfully, applying to database...")
	if _, err := db.Exec(string(schema)); err != nil {
		log.Printf("Error applying schema: %v", err)
		log.Printf("Attempting to apply migration for existing database...")

		// Try to apply migration for existing database
		migrationPath := "schema/migrate_to_simple.sql"
		log.Printf("Attempting to read migration from: %s", migrationPath)

		migration, err := os.ReadFile(migrationPath)
		if err != nil {
			log.Printf("Error reading migration from %s: %v", migrationPath, err)
			log.Printf("Attempting to read migration from absolute path...")
			absMigrationPath := filepath.Join(getCurrentDir(), "schema", "migrate_to_simple.sql")
			log.Printf("Trying absolute path: %s", absMigrationPath)
			migration, err = os.ReadFile(absMigrationPath)
			if err != nil {
				log.Fatal("Error reading migration from absolute path:", err)
			}
		}

		log.Printf("Migration file loaded successfully, applying to database...")
		if _, err := db.Exec(string(migration)); err != nil {
			log.Fatal("Error applying migration:", err)
		}
		log.Printf("Database migration applied successfully")
	} else {
		log.Printf("Database schema applied successfully")
	}

	// Apply emails schema
	log.Printf("Loading emails schema...")
	emailsSchemaPath := "schema/emails.sql"
	log.Printf("Attempting to read emails schema from: %s", emailsSchemaPath)

	emailsSchema, err := os.ReadFile(emailsSchemaPath)
	if err != nil {
		log.Printf("Error reading emails schema from %s: %v", emailsSchemaPath, err)
		log.Printf("Attempting to read emails schema from absolute path...")
		absEmailsPath := filepath.Join(getCurrentDir(), "schema", "emails.sql")
		log.Printf("Trying absolute path: %s", absEmailsPath)
		emailsSchema, err = os.ReadFile(absEmailsPath)
		if err != nil {
			log.Fatal("Error reading emails schema from absolute path:", err)
		}
	}

	log.Printf("Emails schema file loaded successfully, applying to database...")
	if _, err := db.Exec(string(emailsSchema)); err != nil {
		log.Printf("Error applying emails schema: %v", err)
		log.Printf("Emails table may already exist or have incompatible structure")
	} else {
		log.Printf("Emails schema applied successfully")
	}

	// Initialize per-IP rate limiter for signup and login
	signupLoginLimiter := NewIPRateLimitMiddleware(5, time.Minute)

	// Set up router
	r := mux.NewRouter()

	// Apply middleware BEFORE route registration
	r.Use(srv.corsMiddleware)
	r.Use(srv.secureHeadersMiddleware)

	log.Printf("Registering /ping endpoint")
	r.HandleFunc("/ping", srv.pingHandler).Methods("GET")

	log.Printf("Registering /health endpoint")
	r.HandleFunc("/health", srv.healthHandler).Methods("GET")

	// Wrap /login and /signup with rate limit middleware
	log.Printf("Registering /login endpoint")
	r.Handle("/login", signupLoginLimiter.Middleware(loginHandlerFactory(db))).Methods("POST")
	log.Printf("Registering /api/auth/login endpoint")
	r.Handle("/api/auth/login", signupLoginLimiter.Middleware(http.HandlerFunc(srv.loginHandler))).Methods("POST")
	log.Printf("Registering /api/auth/signup endpoint")
	r.Handle("/api/auth/signup", signupLoginLimiter.Middleware(signupHandlerFactory(db))).Methods("POST")
	log.Printf("Registering /api/auth/verify-totp endpoint")
	r.HandleFunc("/api/auth/verify-totp", auth.VerifyTotpHandler(db)).Methods("POST")
	log.Printf("Registering /confirm-fallback endpoint")
	r.HandleFunc("/confirm-fallback", confirmFallbackHandlerFactory(db)).Methods("GET")
	log.Printf("Registering /resend-fallback endpoint")
	r.HandleFunc("/resend-fallback", resendFallbackHandlerFactory(db)).Methods("POST")

	// Protected routes (require JWT authentication)
	log.Printf("Registering /protected-test endpoint")
	r.Handle("/protected-test", jwtMiddleware(http.HandlerFunc(protectedTestHandler))).Methods("GET")

	handler := r

	// Start server
	addr := ":8080"
	log.Printf("Starting API on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal("Server error:", err)
	}
}

// Server struct holds the database connection and rate limiter map for per-IP rate limiting.
type Server struct {
	db         *sql.DB
	rateLimits *sync.Map // IP -> attempt count
}

// pingHandler is a simple health check endpoint for liveness probes.
func (srv *Server) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("pong"))
}

// healthHandler returns a JSON status for monitoring and health checks.
func (srv *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Health handler called from IP: %s", r.RemoteAddr)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// loginHandler handles POST /api/auth/login. It authenticates the user using password and TOTP, and returns a JWT on success.
// If authentication fails, it logs the error and returns an appropriate error response.
func (srv *Server) loginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
		TOTPCode string `json:"totp_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	// Authenticate user
	token, userID, err := auth.Authenticate(srv.db, req.Email, req.Password, req.TOTPCode)
	if err != nil {
		srv.logError(r, req.Email, "Authentication failed")
		http.Error(w, `{"error":"Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	// Respond with JWT
	resp := struct {
		Token  string `json:"token"`
		UserID string `json:"user_id"`
	}{Token: token, UserID: userID}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// corsMiddleware restricts allowed origins and sets CORS headers for security.
func (srv *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Define allowed origins
		allowedOrigins := []string{
			"http://localhost:3000",
			"https://securesystem.email",
		}

		origin := r.Header.Get("Origin")
		allowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				allowed = true
				break
			}
		}

		if allowed {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		// Handle preflight OPTIONS requests
		if r.Method == "OPTIONS" {
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
			w.Header().Set("Access-Control-Allow-Credentials", "true")
			w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
			w.WriteHeader(http.StatusOK)
			return
		}

		// Set CORS headers for actual requests
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")

		next.ServeHTTP(w, r)
	})
}

// secureHeadersMiddleware sets HTTP security headers to protect against common web vulnerabilities.
func (srv *Server) secureHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// CORS-friendly security headers
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

// logError writes authentication and security errors to a log file for audit and monitoring.
func (srv *Server) logError(r *http.Request, email, msg string) {
	logPath := os.Getenv("LOG_FILE")
	if logPath == "" {
		logPath = "/var/log/api.log"
	}
	f, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println("Error opening log file:", err)
		return
	}
	defer f.Close()
	logger := log.New(f, "", log.LstdFlags)
	logger.Printf("Error: %s, Email: %s, IP: %s", msg, email, r.RemoteAddr)
}

// testR2Connection verifies connectivity to Cloudflare R2 storage for future encrypted email storage.
func testR2Connection() error {
	// Get R2 credentials from environment
	accessKey := os.Getenv("CLOUDFLARE_R2_ACCESS_KEY")
	secretKey := os.Getenv("CLOUDFLARE_R2_SECRET_KEY")
	bucket := os.Getenv("CLOUDFLARE_R2_BUCKET")
	endpoint := os.Getenv("CLOUDFLARE_R2_ENDPOINT")

	if accessKey == "" || secretKey == "" || bucket == "" || endpoint == "" {
		return fmt.Errorf("R2 credentials not configured")
	}

	// Create AWS session for R2
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("auto"),
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endpoint),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create R2 session: %v", err)
	}

	// Create S3 client
	s3Client := s3.New(sess)

	// Test upload
	testContent := "Hello from Secure Email MVP!"
	testKey := "test-upload.txt"

	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(testKey),
		Body:   strings.NewReader(testContent),
	})
	if err != nil {
		return fmt.Errorf("failed to upload test file: %v", err)
	}

	// Test download
	result, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(testKey),
	})
	if err != nil {
		return fmt.Errorf("failed to download test file: %v", err)
	}
	defer result.Body.Close()

	// Clean up test file
	_, err = s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(testKey),
	})
	if err != nil {
		log.Printf("Warning: failed to delete test file: %v", err)
	}

	log.Printf("R2 connection test successful - bucket: %s", bucket)
	return nil
}

// getCurrentDir returns the current working directory for logging and file path resolution.
func getCurrentDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "unknown"
	}
	return dir
}

// JWT middleware implementation
// jwtMiddleware validates JWT tokens for protected endpoints and injects the user's email into the request context.
func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"error":"Invalid or missing token"}`, http.StatusUnauthorized)
			return
		}

		// Check Bearer format
		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			http.Error(w, `{"error":"Invalid or missing token"}`, http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[7:]
		if tokenString == "" {
			http.Error(w, `{"error":"Invalid or missing token"}`, http.StatusUnauthorized)
			return
		}

		// Validate JWT token
		claims, err := auth.ParseJWT(tokenString)
		if err != nil {
			http.Error(w, `{"error":"Invalid or missing token"}`, http.StatusUnauthorized)
			return
		}

		// Set email in context using UserEmailKey
		ctx := context.WithValue(r.Context(), UserEmailKey, claims.Subject)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// Protected test handler
// protectedTestHandler is a sample protected endpoint that requires JWT authentication.
func protectedTestHandler(w http.ResponseWriter, r *http.Request) {
	email, ok := GetUserEmailFromContext(r)
	if !ok {
		http.Error(w, `{"error":"Email not found in context"}`, http.StatusInternalServerError)
		return
	}

	response := ProtectedResponse{
		Email:   email,
		Message: "Access granted",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// contextKey is a type for context keys used in request context.
type contextKey string

// UserEmailKey is the context key for storing the user's email in the request context.
const UserEmailKey contextKey = "email"

// GetUserEmailFromContext extracts the user's email from the request context.
func GetUserEmailFromContext(r *http.Request) (string, bool) {
	email, ok := r.Context().Value(UserEmailKey).(string)
	return email, ok
}

// ProtectedResponse represents the response structure for protected endpoints.
type ProtectedResponse struct {
	Email   string `json:"email"`
	Message string `json:"message"`
}

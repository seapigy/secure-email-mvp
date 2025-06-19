package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"secure-email-mvp/pkg/auth"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/cors"
)

type Server struct {
	db         *sql.DB
	rateLimits *sync.Map // IP -> attempt count
}

func main() {
	// Load .env
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env:", err)
	}

	// Connect to SQLite
	dbPath := os.Getenv("SQLITE_DB")
	if dbPath == "" {
		dbPath = "/var/db/secure-email.db"
	}
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		log.Fatal("Error connecting to database:", err)
	}

	// Initialize server
	srv := &Server{db: db, rateLimits: &sync.Map{}}

	// Apply schema
	schema, err := os.ReadFile("schema/users.sql")
	if err != nil {
		log.Fatal("Error reading schema:", err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		log.Fatal("Error applying schema:", err)
	}

	// Set up router
	r := mux.NewRouter()
	r.HandleFunc("/api/auth/login", srv.loginHandler).Methods("POST")
	r.HandleFunc("/api/auth/signup", auth.SignUpHandler(db)).Methods("POST")
	r.HandleFunc("/api/auth/verify-totp", auth.VerifyTotpHandler(db)).Methods("POST")

	// Apply middleware
	r.Use(srv.rateLimitMiddleware)
	r.Use(srv.secureHeadersMiddleware)

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"http://localhost:3000", "https://secure-email-mvp.netlify.app"},
		AllowedMethods: []string{"POST"},
		AllowedHeaders: []string{"Content-Type"},
	})
	handler := c.Handler(r)

	// Start server
	addr := ":8080"
	log.Printf("Starting API on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal("Server error:", err)
	}
}

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

func (srv *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip := r.RemoteAddr
		count, _ := srv.rateLimits.LoadOrStore(ip, 0)
		if count.(int) >= 10 {
			http.Error(w, `{"error":"Too many requests"}`, http.StatusTooManyRequests)
			return
		}
		srv.rateLimits.Store(ip, count.(int)+1)
		go func() {
			time.Sleep(time.Minute)
			srv.rateLimits.Delete(ip)
		}()
		next.ServeHTTP(w, r)
	})
}

func (srv *Server) secureHeadersMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		next.ServeHTTP(w, r)
	})
}

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

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
	"strconv"
	"strings"
	"sync"
	"time"

	"secure-email-mvp/pkg/auth"
	"secure-email-mvp/pkg/cleanup"
	"secure-email-mvp/pkg/iptracking"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"
)

// main is the entry point for the Secure Email API server. It initializes the application by:
// 1. Loading environment variables from .env file
// 2. Establishing database connection with SQLite
// 3. Testing Cloudflare R2 storage connectivity
// 4. Applying database schema migrations (including all security features)
// 5. Setting up HTTP server with middleware (CORS, security headers, rate limiting)
// 6. Registering all API endpoints for authentication, email operations, MFA, and notifications
// 7. Starting the HTTP server on the configured port
//
// Security Features Implemented:
// - Multi-Factor Authentication (TOTP + Email-based)
// - Enhanced Geolocation Verification (City + Country)
// - Per-email Password Protection with Argon2id
// - Brute Force Protection (Per-email + IP-based)
// - Access Notification System
// - Self-Destruct After Failed Attempts
// - Session Management and Rate Limiting
func main() {
	// Initialize logging and load environment configuration
	log.Printf("Starting Secure Email API server...")
	log.Printf("Current working directory: %s", getCurrentDir())

	// Load environment variables from .env file with fallback paths
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

	// Validate critical environment variables
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Printf("Warning: JWT_SECRET not found in environment, using default")
	} else {
		log.Printf("JWT_SECRET loaded successfully")
	}

	// Initialize SQLite database connection
	dbPath := os.Getenv("SQLITE_DB")
	if dbPath == "" {
		dbPath = "/var/db/secure-email.db"
	}
	log.Printf("Attempting to connect to database at: %s", dbPath)

	// Ensure database directory exists with proper permissions
	dbDir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		log.Printf("Warning: Could not create database directory %s: %v", dbDir, err)
	}

	// Open database connection using modernc.org/sqlite driver
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal("Error opening database:", err)
	}
	defer db.Close()

	// Test database connectivity
	log.Printf("Testing database connection...")
	if err := db.Ping(); err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	log.Printf("Database connection successful")

	// Test Cloudflare R2 storage connectivity for encrypted email content
	if err := testR2Connection(); err != nil {
		log.Printf("Warning: R2 connection test failed: %v", err)
	} else {
		log.Printf("R2 connection test passed")
	}

	// Initialize server instance with database and rate limiting
	srv := &Server{db: db, rateLimits: &sync.Map{}}

	// =============================================================================
	// DATABASE MIGRATIONS - SECURITY FEATURES
	// =============================================================================
	// Apply database schema with comprehensive error handling
	// All migrations include fallback paths and error handling for production deployment
	//
	// Migration Order:
	// 1. Core schema (users, emails)
	// 2. Failed attempts tracking
	// 3. Geolocation restrictions
	// 4. Multi-Factor Authentication (MFA)
	// 5. Simple geolocation (Micro-Iteration 4.10)
	// 6. Brute force protection (Micro-Iteration 4.12)
	// 7. IP tracking (Micro-Iteration 4.13)
	// 8. Email password protection (Micro-Iteration 4.14)
	// 9. Enhanced geolocation (Micro-Iteration 4.15)
	// 10. Notification system (Micro-Iteration 4.17)
	// =============================================================================
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

		// Apply migration for existing database if schema application fails
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

	// Apply failed attempts migration
	log.Printf("Loading failed attempts migration...")
	failedAttemptsMigrationPath := "schema/migrate_add_failed_attempts.sql"
	log.Printf("Attempting to read failed attempts migration from: %s", failedAttemptsMigrationPath)

	failedAttemptsMigration, err := os.ReadFile(failedAttemptsMigrationPath)
	if err != nil {
		log.Printf("Error reading failed attempts migration from %s: %v", failedAttemptsMigrationPath, err)
		log.Printf("Attempting to read failed attempts migration from absolute path...")
		absFailedAttemptsPath := filepath.Join(getCurrentDir(), "schema", "migrate_add_failed_attempts.sql")
		log.Printf("Trying absolute path: %s", absFailedAttemptsPath)
		failedAttemptsMigration, err = os.ReadFile(absFailedAttemptsPath)
		if err != nil {
			log.Printf("Error reading failed attempts migration from absolute path: %v", err)
		} else {
			log.Printf("Failed attempts migration file loaded successfully, applying to database...")
			if _, err := db.Exec(string(failedAttemptsMigration)); err != nil {
				log.Printf("Error applying failed attempts migration: %v", err)
				log.Printf("Failed attempts migration may already be applied or have incompatible structure")
			} else {
				log.Printf("Failed attempts migration applied successfully")
			}
		}
	} else {
		log.Printf("Failed attempts migration file loaded successfully, applying to database...")
		if _, err := db.Exec(string(failedAttemptsMigration)); err != nil {
			log.Printf("Error applying failed attempts migration: %v", err)
			log.Printf("Failed attempts migration may already be applied or have incompatible structure")
		} else {
			log.Printf("Failed attempts migration applied successfully")
		}
	}

	// Apply fail_count migration
	log.Printf("Loading fail_count migration...")
	failCountMigrationPath := "schema/migrate_add_fail_count.sql"
	log.Printf("Attempting to read fail_count migration from: %s", failCountMigrationPath)

	failCountMigration, err := os.ReadFile(failCountMigrationPath)
	if err != nil {
		log.Printf("Error reading fail_count migration from %s: %v", failCountMigrationPath, err)
		log.Printf("Attempting to read fail_count migration from absolute path...")
		absFailCountPath := filepath.Join(getCurrentDir(), "schema", "migrate_add_fail_count.sql")
		log.Printf("Trying absolute path: %s", absFailCountPath)
		failCountMigration, err = os.ReadFile(absFailCountPath)
		if err != nil {
			log.Printf("Error reading fail_count migration from absolute path: %v", err)
		} else {
			log.Printf("Fail_count migration file loaded successfully, applying to database...")
			if _, err := db.Exec(string(failCountMigration)); err != nil {
				log.Printf("Error applying fail_count migration: %v", err)
				log.Printf("Fail_count migration may already be applied or have incompatible structure")
			} else {
				log.Printf("Fail_count migration applied successfully")
			}
		}
	} else {
		log.Printf("Fail_count migration file loaded successfully, applying to database...")
		if _, err := db.Exec(string(failCountMigration)); err != nil {
			log.Printf("Error applying fail_count migration: %v", err)
			log.Printf("Fail_count migration may already be applied or have incompatible structure")
		} else {
			log.Printf("Fail_count migration applied successfully")
		}
	}

	// Apply geolocation restrictions migration
	log.Printf("Loading geolocation restrictions migration...")
	geolocationMigrationPath := "schema/migrate_add_geolocation_restrictions.sql"
	log.Printf("Attempting to read geolocation restrictions migration from: %s", geolocationMigrationPath)

	geolocationMigration, err := os.ReadFile(geolocationMigrationPath)
	if err != nil {
		log.Printf("Error reading geolocation restrictions migration from %s: %v", geolocationMigrationPath, err)
		log.Printf("Attempting to read geolocation restrictions migration from absolute path...")
		absGeolocationPath := filepath.Join(getCurrentDir(), "schema", "migrate_add_geolocation_restrictions.sql")
		log.Printf("Trying absolute path: %s", absGeolocationPath)
		geolocationMigration, err = os.ReadFile(absGeolocationPath)
		if err != nil {
			log.Printf("Error reading geolocation restrictions migration from absolute path: %v", err)
		} else {
			log.Printf("Geolocation restrictions migration file loaded successfully, applying to database...")
			if _, err := db.Exec(string(geolocationMigration)); err != nil {
				log.Printf("Error applying geolocation restrictions migration: %v", err)
				log.Printf("Geolocation restrictions migration may already be applied or have incompatible structure")
			} else {
				log.Printf("Geolocation restrictions migration applied successfully")
			}
		}
	} else {
		log.Printf("Geolocation restrictions migration file loaded successfully, applying to database...")
		if _, err := db.Exec(string(geolocationMigration)); err != nil {
			log.Printf("Error applying geolocation restrictions migration: %v", err)
			log.Printf("Geolocation restrictions migration may already be applied or have incompatible structure")
		} else {
			log.Printf("Geolocation restrictions migration applied successfully")
		}
	}

	// Apply MFA migration
	log.Printf("Loading MFA migration...")
	mfaMigrationPath := "schema/migrate_add_mfa_fields.sql"
	log.Printf("Attempting to read MFA migration from: %s", mfaMigrationPath)

	mfaMigration, err := os.ReadFile(mfaMigrationPath)
	if err != nil {
		log.Printf("Error reading MFA migration from %s: %v", mfaMigrationPath, err)
		log.Printf("Attempting to read MFA migration from absolute path...")
		absMFAPath := filepath.Join(getCurrentDir(), "schema", "migrate_add_mfa_fields.sql")
		log.Printf("Trying absolute path: %s", absMFAPath)
		mfaMigration, err = os.ReadFile(absMFAPath)
		if err != nil {
			log.Printf("Error reading MFA migration from absolute path: %v", err)
		} else {
			log.Printf("MFA migration file loaded successfully, applying to database...")
			if _, err := db.Exec(string(mfaMigration)); err != nil {
				log.Printf("Error applying MFA migration: %v", err)
				log.Printf("MFA migration may already be applied or have incompatible structure")
			} else {
				log.Printf("MFA migration applied successfully")
			}
		}
	} else {
		log.Printf("MFA migration file loaded successfully, applying to database...")
		if _, err := db.Exec(string(mfaMigration)); err != nil {
			log.Printf("Error applying MFA migration: %v", err)
			log.Printf("MFA migration may already be applied or have incompatible structure")
		} else {
			log.Printf("MFA migration applied successfully")
		}
	}

	// Apply simple geolocation migration (Micro-Iteration 4.10)
	log.Printf("Loading simple geolocation migration...")
	simpleGeoMigrationPath := "schema/migrate_add_simple_geolocation.sql"
	log.Printf("Attempting to read simple geolocation migration from: %s", simpleGeoMigrationPath)

	simpleGeoMigration, err := os.ReadFile(simpleGeoMigrationPath)
	if err != nil {
		log.Printf("Error reading simple geolocation migration from %s: %v", simpleGeoMigrationPath, err)
		log.Printf("Attempting to read simple geolocation migration from absolute path...")
		absSimpleGeoPath := filepath.Join(getCurrentDir(), "schema", "migrate_add_simple_geolocation.sql")
		log.Printf("Trying absolute path: %s", absSimpleGeoPath)
		simpleGeoMigration, err = os.ReadFile(absSimpleGeoPath)
		if err != nil {
			log.Printf("Error reading simple geolocation migration from absolute path: %v", err)
		} else {
			log.Printf("Simple geolocation migration file loaded successfully, applying to database...")
			if _, err := db.Exec(string(simpleGeoMigration)); err != nil {
				log.Printf("Error applying simple geolocation migration: %v", err)
				log.Printf("Simple geolocation migration may already be applied or have incompatible structure")
			} else {
				log.Printf("Simple geolocation migration applied successfully")
			}
		}
	} else {
		log.Printf("Simple geolocation migration file loaded successfully, applying to database...")
		if _, err := db.Exec(string(simpleGeoMigration)); err != nil {
			log.Printf("Error applying simple geolocation migration: %v", err)
			log.Printf("Simple geolocation migration may already be applied or have incompatible structure")
		} else {
			log.Printf("Simple geolocation migration applied successfully")
		}
	}

	// Apply brute-force protection migration (Micro-Iteration 4.12)
	log.Printf("Loading brute-force protection migration...")
	bruteForceMigrationPath := "schema/migrate_add_brute_force_protection.sql"
	log.Printf("Attempting to read brute-force protection migration from: %s", bruteForceMigrationPath)

	bruteForceMigration, err := os.ReadFile(bruteForceMigrationPath)
	if err != nil {
		log.Printf("Error reading brute-force protection migration from %s: %v", bruteForceMigrationPath, err)
		log.Printf("Attempting to read brute-force protection migration from absolute path...")
		absBruteForcePath := filepath.Join(getCurrentDir(), "schema", "migrate_add_brute_force_protection.sql")
		log.Printf("Trying absolute path: %s", absBruteForcePath)
		bruteForceMigration, err = os.ReadFile(absBruteForcePath)
		if err != nil {
			log.Printf("Error reading brute-force protection migration from absolute path: %v", err)
		} else {
			log.Printf("Brute-force protection migration file loaded successfully, applying to database...")
			if _, err := db.Exec(string(bruteForceMigration)); err != nil {
				log.Printf("Error applying brute-force protection migration: %v", err)
				log.Printf("Brute-force protection migration may already be applied or have incompatible structure")
			} else {
				log.Printf("Brute-force protection migration applied successfully")
			}
		}
	} else {
		log.Printf("Brute-force protection migration file loaded successfully, applying to database...")
		if _, err := db.Exec(string(bruteForceMigration)); err != nil {
			log.Printf("Error applying brute-force protection migration: %v", err)
			log.Printf("Brute-force protection migration may already be applied or have incompatible structure")
		} else {
			log.Printf("Brute-force protection migration applied successfully")
		}
	}

	// Apply IP tracking migration (Micro-Iteration 4.13)
	log.Printf("Loading IP tracking migration...")
	ipTrackingMigrationPath := "schema/migrate_add_ip_tracking.sql"
	log.Printf("Attempting to read IP tracking migration from: %s", ipTrackingMigrationPath)

	ipTrackingMigration, err := os.ReadFile(ipTrackingMigrationPath)
	if err != nil {
		log.Printf("Error reading IP tracking migration from %s: %v", ipTrackingMigrationPath, err)
		log.Printf("Attempting to read IP tracking migration from absolute path...")
		absIPTrackingPath := filepath.Join(getCurrentDir(), "schema", "migrate_add_ip_tracking.sql")
		log.Printf("Trying absolute path: %s", absIPTrackingPath)
		ipTrackingMigration, err = os.ReadFile(absIPTrackingPath)
		if err != nil {
			log.Printf("Error reading IP tracking migration from absolute path: %v", err)
		} else {
			log.Printf("IP tracking migration file loaded successfully, applying to database...")
			if _, err := db.Exec(string(ipTrackingMigration)); err != nil {
				log.Printf("Error applying IP tracking migration: %v", err)
				log.Printf("IP tracking migration may already be applied or have incompatible structure")
			} else {
				log.Printf("IP tracking migration applied successfully")
			}
		}
	} else {
		log.Printf("IP tracking migration file loaded successfully, applying to database...")
		if _, err := db.Exec(string(ipTrackingMigration)); err != nil {
			log.Printf("Error applying IP tracking migration: %v", err)
			log.Printf("IP tracking migration may already be applied or have incompatible structure")
		} else {
			log.Printf("IP tracking migration applied successfully")
		}
	}

	// Initialize IP tracking service and run cleanup
	ipTrackingService := iptracking.NewIPTrackingService(db)
	if err := ipTrackingService.CleanupOldRecords(); err != nil {
		log.Printf("Warning: Failed to cleanup old IP records: %v", err)
	} else {
		log.Printf("IP tracking cleanup completed successfully")
	}

	// Apply password protection migration (Micro-Iteration 4.14)
	log.Printf("Loading password protection migration...")
	passwordProtectionMigrationPath := "schema/migrate_add_email_password_protection.sql"
	log.Printf("Attempting to read password protection migration from: %s", passwordProtectionMigrationPath)

	passwordProtectionMigration, err := os.ReadFile(passwordProtectionMigrationPath)
	if err != nil {
		log.Printf("Error reading password protection migration from %s: %v", passwordProtectionMigrationPath, err)
		log.Printf("Attempting to read password protection migration from absolute path...")
		absPasswordProtectionPath := filepath.Join(getCurrentDir(), "schema", "migrate_add_email_password_protection.sql")
		log.Printf("Trying absolute path: %s", absPasswordProtectionPath)
		passwordProtectionMigration, err = os.ReadFile(absPasswordProtectionPath)
		if err != nil {
			log.Printf("Error reading password protection migration from absolute path: %v", err)
		} else {
			log.Printf("Password protection migration file loaded successfully, applying to database...")
			if _, err := db.Exec(string(passwordProtectionMigration)); err != nil {
				log.Printf("Error applying password protection migration: %v", err)
				log.Printf("Password protection migration may already be applied or have incompatible structure")
			} else {
				log.Printf("Password protection migration applied successfully")
			}
		}
	} else {
		log.Printf("Password protection migration file loaded successfully, applying to database...")
		if _, err := db.Exec(string(passwordProtectionMigration)); err != nil {
			log.Printf("Error applying password protection migration: %v", err)
			log.Printf("Password protection migration may already be applied or have incompatible structure")
		} else {
			log.Printf("Password protection migration applied successfully")
		}
	}

	// Apply enhanced geolocation verification migration (Micro-Iteration 4.15)
	log.Printf("Loading enhanced geolocation verification migration...")
	geoVerificationMigrationPath := "schema/migrate_add_city_country_verification.sql"
	log.Printf("Attempting to read enhanced geolocation verification migration from: %s", geoVerificationMigrationPath)

	geoVerificationMigration, err := os.ReadFile(geoVerificationMigrationPath)
	if err != nil {
		log.Printf("Error reading enhanced geolocation verification migration from %s: %v", geoVerificationMigrationPath, err)
		log.Printf("Attempting to read enhanced geolocation verification migration from absolute path...")
		absGeoVerificationPath := filepath.Join(getCurrentDir(), "schema", "migrate_add_city_country_verification.sql")
		log.Printf("Trying absolute path: %s", absGeoVerificationPath)
		geoVerificationMigration, err = os.ReadFile(absGeoVerificationPath)
		if err != nil {
			log.Printf("Error reading enhanced geolocation verification migration from absolute path: %v", err)
		} else {
			log.Printf("Enhanced geolocation verification migration file loaded successfully, applying to database...")
			if _, err := db.Exec(string(geoVerificationMigration)); err != nil {
				log.Printf("Error applying enhanced geolocation verification migration: %v", err)
				log.Printf("Enhanced geolocation verification migration may already be applied or have incompatible structure")
			} else {
				log.Printf("Enhanced geolocation verification migration applied successfully")
			}
		}
	} else {
		log.Printf("Enhanced geolocation verification migration file loaded successfully, applying to database...")
		if _, err := db.Exec(string(geoVerificationMigration)); err != nil {
			log.Printf("Error applying enhanced geolocation verification migration: %v", err)
			log.Printf("Enhanced geolocation verification migration may already be applied or have incompatible structure")
		} else {
			log.Printf("Enhanced geolocation verification migration applied successfully")
		}
	}

	// Apply notification system migration (Micro-Iteration 4.17)
	log.Printf("Loading notification system migration...")
	notificationMigrationPath := "schema/migrate_add_notification_system.sql"
	log.Printf("Attempting to read notification system migration from: %s", notificationMigrationPath)

	notificationMigration, err := os.ReadFile(notificationMigrationPath)
	if err != nil {
		log.Printf("Error reading notification system migration from %s: %v", notificationMigrationPath, err)
		log.Printf("Attempting to read notification system migration from absolute path...")
		absNotificationPath := filepath.Join(getCurrentDir(), "schema", "migrate_add_notification_system.sql")
		log.Printf("Trying absolute path: %s", absNotificationPath)
		notificationMigration, err = os.ReadFile(absNotificationPath)
		if err != nil {
			log.Printf("Error reading notification system migration from absolute path: %v", err)
		} else {
			log.Printf("Notification system migration file loaded successfully, applying to database...")
			if _, err := db.Exec(string(notificationMigration)); err != nil {
				log.Printf("Error applying notification system migration: %v", err)
				log.Printf("Notification system migration may already be applied or have incompatible structure")
			} else {
				log.Printf("Notification system migration applied successfully")
			}
		}
	} else {
		log.Printf("Notification system migration file loaded successfully, applying to database...")
		if _, err := db.Exec(string(notificationMigration)); err != nil {
			log.Printf("Error applying notification system migration: %v", err)
			log.Printf("Notification system migration may already be applied or have incompatible structure")
		} else {
			log.Printf("Notification system migration applied successfully")
		}
	}

	// Initialize per-IP rate limiter for signup and login
	signupLoginLimiter := NewIPRateLimitMiddleware(5, time.Minute)

	// =============================================================================
	// HTTP SERVER SETUP - API ENDPOINTS
	// =============================================================================
	// Set up router with comprehensive security middleware
	// All endpoints are protected with JWT middleware and rate limiting
	//
	// Endpoint Categories:
	// - Authentication: Login, signup, TOTP, refresh, logout
	// - Email Operations: Send, get, list, view, delete
	// - Multi-Factor Authentication: TOTP and email-based MFA
	// - Notifications: Preferences and access event history
	// - Admin: Cleanup statistics and manual triggers
	// - Health: System health monitoring
	// =============================================================================
	
	// Set up router
	r := mux.NewRouter()

	// Apply middleware BEFORE route registration
	r.Use(srv.corsMiddleware)
	r.Use(srv.secureHeadersMiddleware)

	log.Printf("Registering /ping endpoint")
	r.HandleFunc("/ping", srv.pingHandler).Methods("GET")

	log.Printf("Registering /health endpoint")
	r.HandleFunc("/health", srv.healthHandler).Methods("GET")

	// Apply refresh tokens schema
	log.Printf("Loading refresh tokens schema...")
	refreshTokensSchemaPath := "schema/refresh_tokens.sql"
	log.Printf("Attempting to read refresh tokens schema from: %s", refreshTokensSchemaPath)

	refreshTokensSchema, err := os.ReadFile(refreshTokensSchemaPath)
	if err != nil {
		log.Printf("Error reading refresh tokens schema from %s: %v", refreshTokensSchemaPath, err)
		log.Printf("Attempting to read refresh tokens schema from absolute path...")
		absRefreshTokensPath := filepath.Join(getCurrentDir(), "schema", "refresh_tokens.sql")
		log.Printf("Trying absolute path: %s", absRefreshTokensPath)
		refreshTokensSchema, err = os.ReadFile(absRefreshTokensPath)
		if err != nil {
			log.Printf("Error reading refresh tokens schema from absolute path: %v", err)
		} else {
			log.Printf("Refresh tokens schema file loaded successfully, applying to database...")
			if _, err := db.Exec(string(refreshTokensSchema)); err != nil {
				log.Printf("Error applying refresh tokens schema: %v", err)
				log.Printf("Refresh tokens table may already exist or have incompatible structure")
			} else {
				log.Printf("Refresh tokens schema applied successfully")
			}
		}
	} else {
		log.Printf("Refresh tokens schema file loaded successfully, applying to database...")
		if _, err := db.Exec(string(refreshTokensSchema)); err != nil {
			log.Printf("Error applying refresh tokens schema: %v", err)
			log.Printf("Refresh tokens table may already exist or have incompatible structure")
		} else {
			log.Printf("Refresh tokens schema applied successfully")
		}
	}

	// Wrap /login and /signup with rate limit middleware
	log.Printf("Registering /login endpoint")
	r.Handle("/login", signupLoginLimiter.Middleware(http.HandlerFunc(srv.loginHandler))).Methods("POST")
	log.Printf("Registering /api/auth/login endpoint")
	r.Handle("/api/auth/login", signupLoginLimiter.Middleware(loginHandler(db))).Methods("POST")
	log.Printf("Registering /api/auth/signup endpoint")
	r.Handle("/api/auth/signup", signupLoginLimiter.Middleware(signupHandlerFactory(db))).Methods("POST")
	log.Printf("Registering /api/auth/verify-totp endpoint")
	r.HandleFunc("/api/auth/verify-totp", auth.VerifyTotpHandler(db)).Methods("POST")
	log.Printf("Registering /api/auth/refresh endpoint")
	r.HandleFunc("/api/auth/refresh", refreshHandler(db)).Methods("POST")
	log.Printf("Registering /api/auth/logout endpoint")
	r.HandleFunc("/api/auth/logout", logoutHandler(db)).Methods("POST")
	log.Printf("Registering /api/auth/me endpoint")
	r.Handle("/api/auth/me", jwtMiddleware(meHandler())).Methods("GET")
	log.Printf("Registering /confirm-fallback endpoint")
	r.HandleFunc("/confirm-fallback", confirmFallbackHandlerFactory(db)).Methods("GET")
	log.Printf("Registering /resend-fallback endpoint")
	r.HandleFunc("/resend-fallback", resendFallbackHandlerFactory(db)).Methods("POST")

	// Register protected email endpoints (require JWT authentication)
	log.Printf("Registering /api/email/send endpoint")
	r.Handle("/api/email/send", jwtMiddleware(http.HandlerFunc(srv.sendEmailHandler))).Methods("POST")
	log.Printf("Registering /api/email/get endpoint")
	r.Handle("/api/email/get", jwtMiddleware(http.HandlerFunc(srv.getEmailHandler))).Methods("POST")
	log.Printf("Registering /api/email/list endpoint")
	r.Handle("/api/email/list", jwtMiddleware(http.HandlerFunc(srv.listEmailHandler))).Methods("GET")
	log.Printf("Registering /api/email/view/{id} endpoint")
	r.Handle("/api/email/view/{id}", jwtMiddleware(http.HandlerFunc(srv.viewEmailHandler))).Methods("GET")
	log.Printf("Registering /api/email/{id} endpoint")
	r.Handle("/api/email/{id}", jwtMiddleware(http.HandlerFunc(srv.deleteEmailHandler))).Methods("DELETE")

	// Register MFA endpoints (require JWT authentication)
	log.Printf("Registering /api/mfa/validate endpoint")
	r.Handle("/api/mfa/validate", jwtMiddleware(http.HandlerFunc(srv.validateMFAHandler))).Methods("POST")
	log.Printf("Registering /api/mfa/email-code endpoint")
	r.Handle("/api/mfa/email-code", jwtMiddleware(http.HandlerFunc(srv.generateEmailCodeHandler))).Methods("POST")
	log.Printf("Registering /api/mfa/config/{emailID} endpoint")
	r.Handle("/api/mfa/config/{emailID}", jwtMiddleware(http.HandlerFunc(srv.getMFAConfigHandler))).Methods("GET")

	// Register notification endpoints (require JWT authentication)
	log.Printf("Registering /api/notifications/preferences endpoint")
	r.Handle("/api/notifications/preferences", jwtMiddleware(http.HandlerFunc(srv.getNotificationPreferencesHandler))).Methods("GET")
	r.Handle("/api/notifications/preferences", jwtMiddleware(http.HandlerFunc(srv.updateNotificationPreferencesHandler))).Methods("PUT")
	log.Printf("Registering /api/notifications/history endpoint")
	r.Handle("/api/notifications/history", jwtMiddleware(http.HandlerFunc(srv.getAccessEventHistoryHandler))).Methods("GET")

	// Register admin endpoints (require JWT authentication)
	log.Printf("Registering /admin/email-retention-stats endpoint")
	r.Handle("/admin/email-retention-stats", jwtMiddleware(http.HandlerFunc(srv.adminEmailRetentionStatsHandler))).Methods("GET")
	log.Printf("Registering /admin/manual-cleanup endpoint")
	r.Handle("/admin/manual-cleanup", jwtMiddleware(http.HandlerFunc(srv.adminManualCleanupHandler))).Methods("POST")

	// TEST ONLY: Register self-destruct test endpoint
	if os.Getenv("SIMULATE_SELF_DESTRUCT") == "1" {
		log.Printf("Registering /test/self-destruct endpoint (TEST ONLY)")
		r.HandleFunc("/test/self-destruct", srv.testSelfDestructHandler).Methods("POST")
	}

	// Protected routes (require JWT authentication)
	log.Printf("Registering /protected-test endpoint")
	r.Handle("/protected-test", jwtMiddleware(http.HandlerFunc(protectedTestHandler))).Methods("GET")

	// Initialize and start email cleanup worker
	log.Printf("Initializing email cleanup worker...")
	cleanupIntervalStr := os.Getenv("EMAIL_CLEANUP_INTERVAL_MINUTES")
	if cleanupIntervalStr == "" {
		cleanupIntervalStr = "15" // Default to 15 minutes
	}
	cleanupInterval, err := strconv.Atoi(cleanupIntervalStr)
	if err != nil {
		log.Printf("Invalid EMAIL_CLEANUP_INTERVAL_MINUTES, using default 15 minutes")
		cleanupInterval = 15
	}

	cleanupWorker, err := cleanup.NewEmailCleanupWorkerWithDB(db, cleanupInterval)
	if err != nil {
		log.Printf("Warning: Failed to initialize cleanup worker: %v", err)
		log.Printf("Email cleanup will not be available")
	} else {
		log.Printf("Starting email cleanup worker with interval: %d minutes", cleanupInterval)
		cleanupWorker.Start()
		defer cleanupWorker.Stop()
	}

	handler := r

	// Start server
	addr := ":8080"
	log.Printf("Starting API on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal("Server error:", err)
	}
}

// Server struct holds the database connection and rate limiter map for per-IP rate limiting.
// Server represents the main API server with all security features
type Server struct {
	db         *sql.DB                    // SQLite database connection
	rateLimits *sync.Map                  // IP -> attempt count for rate limiting
	// Security features implemented:
	// - Multi-Factor Authentication (TOTP + Email-based)
	// - Enhanced Geolocation Verification (City + Country)
	// - Per-email Password Protection with Argon2id
	// - Brute Force Protection (Per-email + IP-based)
	// - Access Notification System
	// - Self-Destruct After Failed Attempts
	// - Session Management and Rate Limiting
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
// jwtMiddleware validates JWT tokens for protected endpoints and injects the user's user_id into the request context.
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

		// Set user_id in context using UserIDKey
		ctx := context.WithValue(r.Context(), UserIDKey, claims.Subject)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// Protected test handler
// protectedTestHandler is a sample protected endpoint that requires JWT authentication.
func protectedTestHandler(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r)
	if !ok {
		http.Error(w, `{"error":"User ID not found in context"}`, http.StatusInternalServerError)
		return
	}

	response := ProtectedResponse{
		Email:   userID, // Using userID as email for backward compatibility
		Message: "Access granted",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// contextKey is a type for context keys used in request context.
type contextKey string

// UserIDKey is the context key for storing the user's ID in the request context.
const UserIDKey contextKey = "user_id"

// GetUserIDFromContext extracts the user's ID from the request context.
func GetUserIDFromContext(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	return userID, ok
}

// GetUserID is a helper function for extracting user_id from context
func GetUserID(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return "", fmt.Errorf("user_id not found in context")
	}
	return userID, nil
}

// ProtectedResponse represents the response structure for protected endpoints.
type ProtectedResponse struct {
	Email   string `json:"email"`
	Message string `json:"message"`
}

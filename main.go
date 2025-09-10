package main

import (
	"log"
	"net/http"
	"os"

	"secure-email/handlers"

	"github.com/gorilla/mux"
)

func main() {
	// Create router
	r := mux.NewRouter()

	// Register signup route
	r.HandleFunc("/api/signup", handlers.SignupHandler).Methods("POST")

	// Health check endpoint
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// Get port from environment
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

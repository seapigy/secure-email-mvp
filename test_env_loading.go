package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	fmt.Println("=== Testing Environment Variable Loading ===")

	// Try to load .env file
	fmt.Println("Attempting to load .env file...")
	if err := godotenv.Load(); err != nil {
		fmt.Printf("❌ Error loading .env file: %v\n", err)
	} else {
		fmt.Println("✅ Successfully loaded .env file")
	}

	// Check R2 environment variables
	r2Vars := []string{
		"R2_ACCESS_KEY_ID",
		"R2_SECRET_ACCESS_KEY",
		"R2_BUCKET",
		"R2_ENDPOINT",
		"R2_REGION",
	}

	fmt.Println("\nChecking R2 environment variables:")
	for _, varName := range r2Vars {
		value := os.Getenv(varName)
		if value != "" {
			if varName == "R2_SECRET_ACCESS_KEY" {
				fmt.Printf("  ✅ %s = %s...\n", varName, value[:8])
			} else {
				fmt.Printf("  ✅ %s = %s\n", varName, value)
			}
		} else {
			fmt.Printf("  ❌ %s = NOT SET\n", varName)
		}
	}

	// Check current working directory
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ Error getting current directory: %v\n", err)
	} else {
		fmt.Printf("\nCurrent working directory: %s\n", cwd)
	}

	// Check if .env file exists
	if _, err := os.Stat(".env"); os.IsNotExist(err) {
		fmt.Println("❌ .env file does not exist")
	} else {
		fmt.Println("✅ .env file exists")
	}
}

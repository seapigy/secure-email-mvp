package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/pquerna/otp/totp"
)

func main() {
	// Generate fresh TOTP code
	totpSecret := "JBSWY3DPEHPK3PXP"
	code, err := totp.GenerateCode(totpSecret, time.Now())
	if err != nil {
		fmt.Printf("Error generating TOTP: %v\n", err)
		return
	}

	// Test credentials with fresh TOTP code
	loginData := map[string]string{
		"email":     "test@securesystem.email",
		"password":  "testpassword123",
		"totp_code": code,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(loginData)
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}

	fmt.Println("🔧 Testing Login with Fresh TOTP Code")
	fmt.Println("================================================")
	fmt.Printf("Email: %s\n", loginData["email"])
	fmt.Printf("Password: %s\n", loginData["password"])
	fmt.Printf("TOTP Code: %s (generated at %s)\n", loginData["totp_code"], time.Now().Format("15:04:05"))

	// Make HTTP request
	resp, err := http.Post("http://localhost:8080/api/auth/login", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return
	}

	fmt.Printf("\nStatus: %d\n", resp.StatusCode)
	fmt.Printf("Response: %s\n", string(body))

	if resp.StatusCode == 200 {
		fmt.Println("\n✅ Login successful! You can now test SES email sending.")
		fmt.Println("\n📧 Next Steps:")
		fmt.Println("1. Use the access_token from the response")
		fmt.Println("2. Send test email to cpigusch@gmail.com")
		fmt.Println("3. Verify email headers (SPF, DKIM, DMARC)")
	} else {
		fmt.Println("\n❌ Login failed. Check the response for details.")
	}
}


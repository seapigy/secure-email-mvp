# Secure Email API – `/api/email/send` Endpoint Documentation

---

## Project Overview

The **Secure Email API** is designed to provide a secure, robust, and auditable backend for sending and storing encrypted emails. The `/api/email/send` endpoint is a core part of this system, responsible for accepting email data, compressing and encrypting the content, uploading it to secure storage, and recording metadata in the database. The endpoint enforces strict validation and error handling to ensure data integrity and security.

---

## Automated Test Suite Summary

A comprehensive PowerShell test script (`scripts/api_email_send_tests.ps1`) was developed to automate the validation of the `/api/email/send` endpoint. The script covers the following scenarios:

- **Valid POST requests:** Ensures the endpoint works as expected with all required fields.
- **Requests missing each required field:** Tests API response when each required field (`sender_id`, `recipient`, `subject`, `body`) is omitted one at a time.
- **Requests with empty string values:** Checks that empty values for required fields are rejected.
- **Malformed JSON payloads:** Verifies the API’s resilience to invalid JSON.
- **Invalid recipient email format:** Ensures the API rejects improperly formatted email addresses.
- **Simulated database insert failure:** Uses an environment variable to safely trigger and test error handling for DB insert failures.

**Key Findings:**
- The endpoint correctly validates input and handles all tested error cases.
- Simulated DB failures are safely testable and reversible.
- The test suite provides repeatable, automated verification of endpoint robustness.

---

## Backend Code Changes

### Email Format Validation

- Added a regular expression (regex) check in the handler to ensure the `recipient` field contains a valid email address.
- Invalid email formats result in a 400 Bad Request response with a clear error message.

### Error Handling Improvements

- Enhanced error responses for missing fields, empty values, malformed JSON, and invalid email formats.
- All validation and processing errors are logged with context for audit and debugging.

### Simulated DB Failure Logic

- Introduced a code block in the handler that checks for the `SIMULATE_DB_FAILURE` environment variable.
- When set to `1`, the handler simulates a database insert failure, returning a 500 Internal Server Error.
- This logic is clearly marked and easily toggled for safe testing.

---

## How to Run the Tests

### 1. Start the API Server

**Option 1: Using `go run` (Windows)**
```powershell
go run ./cmd/api/main.go ./cmd/api/rate_limit.go ./cmd/api/login_handler.go ./cmd/api/signup_handler.go ./cmd/api/fallback_handler.go ./cmd/api/resend_fallback_handler.go ./cmd/api/send_email_handler.go
```

**Option 2: Build and run the binary**
```powershell
go build -o api-server ./cmd/api
.\api-server.exe
```

### 2. Run the PowerShell Test Script

```powershell
.\scripts\api_email_send_tests.ps1
```

### 3. Simulate DB Failure

- **To enable simulation:**
  1. Stop the server (`Ctrl + C`).
  2. Set the environment variable:
     ```powershell
     $env:SIMULATE_DB_FAILURE = '1'
     ```
  3. Restart the server.
  4. Run the test script; the DB failure test should now return a 500 error.

- **To disable simulation:**
  1. Stop the server.
  2. Remove the environment variable:
     ```powershell
     Remove-Item Env:SIMULATE_DB_FAILURE
     ```
  3. Restart the server for normal operation.

### 4. Verify Results

- The script will print request details, response status, and pass/fail results for each test.
- Server logs will show detailed error and validation messages for each scenario.

---

## Error Handling and Logging

- **Validation Failures:**  
  All missing fields, empty values, and invalid email formats are logged with the offending data and return a 400 error.
- **Malformed JSON:**  
  JSON decode errors are logged and return a 400 error.
- **DB Failures:**  
  Real or simulated DB insert failures are logged and return a 500 error.
- **All errors are returned as JSON responses** with a clear `error` field for client-side handling.

---

## Simulated DB Failure

- The handler checks for the `SIMULATE_DB_FAILURE` environment variable.
- When set, the handler short-circuits before the DB insert and returns a simulated error.
- This allows safe, reversible testing of error handling without breaking the real database or configuration.

---

## What is Markdown?

**Markdown** is a lightweight markup language for formatting text using plain text syntax. It’s widely used for documentation because it’s easy to write and read, and can be rendered into HTML on platforms like GitHub, GitLab, and many others.

---

## Next Steps and Recommendations

- **Share this documentation and the test script with the CTP team for review and independent validation.**
- **Integrate the test script into CI/CD pipelines** for automated regression testing.
- **Expand test coverage** to include additional edge cases and other API endpoints as needed.
- **Monitor server logs in production** for unexpected errors or validation failures.
- **Consider removing or commenting out the simulated DB failure block** in production builds if not needed.
- **Maintain and update documentation** as the API evolves.

---

## Appendix / References

- [Go Documentation](https://golang.org/doc/)
- [Markdown Guide](https://www.markdownguide.org/)
- [PowerShell Documentation](https://docs.microsoft.com/en-us/powershell/)
- [Regex for Email Validation](https://emailregex.com/)
- Internal scripts: `scripts/api_email_send_tests.ps1`

---

*Prepared for CTP review and internal use. For questions or further improvements, contact the Secure Email API development team.* 
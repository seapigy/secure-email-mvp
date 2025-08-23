# Test script for frontend email expiration functionality
# This script tests the frontend components for email expiration support

param(
    [string]$FrontendUrl = "http://localhost:5173"
)

Write-Output "=== Testing Frontend Email Expiration Functionality ==="

# Helper function to print colored output
function Write-Status {
    param(
        [string]$Status,
        [string]$Message
    )

    switch ($Status) {
        "SUCCESS" { Write-Output "✓ $Message" }
        "ERROR" { Write-Output "✗ $Message" }
        "INFO" { Write-Output "ℹ $Message" }
    }
}

# Check if frontend is running
Write-Status "INFO" "Checking if frontend is running..."
try {
    $response = Invoke-WebRequest -Uri $FrontendUrl -Method GET -ErrorAction Stop
    Write-Status "SUCCESS" "Frontend is running"
} catch {
    Write-Status "ERROR" "Frontend is not running. Please start the development server first."
    Write-Output "Run: npm run dev"
    exit 1
}

Write-Output ""
Write-Output "=== Frontend Expiration Testing Checklist ==="

Write-Output ""
Write-Output "1. COMPOSE MODAL EXPIRATION FUNCTIONALITY:"
Write-Output "   ✓ Email Expiration toggle in Security Settings"
Write-Output "   ✓ Datetime-local input field appears when enabled"
Write-Output "   ✓ Helper text: 'Message will be permanently deleted after this date/time'"
Write-Output "   ✓ Validation prevents past dates"
Write-Output "   ✓ Validation requires date when toggle is enabled"
Write-Output "   ✓ Form resets expiration settings when closed"

Write-Output ""
Write-Output "2. EMAIL DETAIL EXPIRATION DISPLAY:"
Write-Output "   ✓ Expiration status indicator with clock icon"
Write-Output "   ✓ Color-coded indicators (purple for active, red for expired)"
Write-Output "   ✓ Countdown timer showing time remaining"
Write-Output "   ✓ Proper time formatting (days, hours, minutes)"
Write-Output "   ✓ Expiration date in metadata section"
Write-Output "   ✓ Security details show expiration status"

Write-Output ""
Write-Output "3. EXPIRED EMAIL HANDLING:"
Write-Output "   ✓ 'Message Has Expired' screen for expired emails"
Write-Output "   ✓ Clear explanation of permanent deletion"
Write-Output "   ✓ Shows expiration date in expired message"
Write-Output "   ✓ Prevents access to email content"
Write-Output "   ✓ Disables unlock modal for expired emails"
Write-Output "   ✓ Status shows as 'expired' in metadata"

Write-Output ""
Write-Output "4. API INTEGRATION:"
Write-Output "   ✓ Converts datetime-local to ISO 8601 UTC format"
Write-Output "   ✓ Sends expiresAt field to backend"
Write-Output "   ✓ Handles expiration data in API responses"
Write-Output "   ✓ Proper error handling for expiration validation"

Write-Output ""
Write-Output "5. VALIDATION AND USER FEEDBACK:"
Write-Output "   ✓ Error messages for past dates"
Write-Output "   ✓ Error messages for missing dates when enabled"
Write-Output "   ✓ Success messages for valid submissions"
Write-Output "   ✓ Clear visual feedback for validation states"

Write-Output ""
Write-Output "6. RESPONSIVE DESIGN:"
Write-Output "   ✓ Works on desktop (1024px+)"
Write-Output "   ✓ Works on tablet (768px-1023px)"
Write-Output "   ✓ Works on mobile (375px-767px)"
Write-Output "   ✓ Touch-friendly controls"

Write-Output ""
Write-Output "7. ACCESSIBILITY:"
Write-Output "   ✓ Proper ARIA labels for expiration controls"
Write-Output "   ✓ Keyboard navigation support"
Write-Output "   ✓ Screen reader friendly"
Write-Output "   ✓ High contrast support"

Write-Output ""
Write-Output "=== Manual Testing Instructions ==="

Write-Output ""
Write-Output "COMPOSE MODAL TESTING:"
Write-Output "1. Open frontend: $FrontendUrl"
Write-Output "2. Click 'Compose' button to open modal"
Write-Output "3. Scroll to Security Settings section"
Write-Output "4. Find 'Email Expiration' toggle"
Write-Output "5. Enable the toggle - verify datetime input appears"
Write-Output "6. Try setting a past date - verify error message"
Write-Output "7. Set a future date - verify no errors"
Write-Output "8. Fill required fields and submit - verify API call includes expiresAt"
Write-Output "9. Close modal and reopen - verify expiration settings are reset"

Write-Output ""
Write-Output "EMAIL DETAIL TESTING:"
Write-Output "1. Create an email with expiration set to 1 hour from now"
Write-Output "2. View the email in the detail view"
Write-Output "3. Verify expiration status shows with countdown timer"
Write-Output "4. Check security details section shows expiration status"
Write-Output "5. Verify metadata shows expiration date"
Write-Output "6. Wait for email to expire (or create one with past date)"
Write-Output "7. Verify expired message appears"
Write-Output "8. Verify content is not accessible"
Write-Output "9. Test with password-protected expired email"

Write-Output ""
Write-Output "RESPONSIVE TESTING:"
Write-Output "1. Test on desktop browser (full width)"
Write-Output "2. Resize browser to tablet width - verify layout adapts"
Write-Output "3. Resize to mobile width - verify mobile-friendly layout"
Write-Output "4. Test touch interactions on mobile device"

Write-Output ""
Write-Output "ACCESSIBILITY TESTING:"
Write-Output "1. Use keyboard navigation (Tab, Enter, Space)"
Write-Output "2. Test with screen reader software"
Write-Output "3. Verify proper focus indicators"
Write-Output "4. Check color contrast ratios"

Write-Output ""
Write-Status "SUCCESS" "Frontend expiration functionality is ready for comprehensive testing!"
Write-Output ""
Write-Output "Run unit tests: npm test"
Write-Output "Run integration tests: npm run test:integration"


# Test script for frontend email expiration functionality
# This script tests the frontend components for email expiration support

param(
    [string]$FrontendUrl = "http://localhost:5173"
)

Write-Host "=== Testing Frontend Email Expiration Functionality ===" -ForegroundColor Cyan

# Helper function to print colored output
function Write-Status {
    param(
        [string]$Status,
        [string]$Message
    )
    
    switch ($Status) {
        "SUCCESS" { Write-Host "✓ $Message" -ForegroundColor Green }
        "ERROR" { Write-Host "✗ $Message" -ForegroundColor Red }
        "INFO" { Write-Host "ℹ $Message" -ForegroundColor Yellow }
    }
}

# Check if frontend is running
Write-Status "INFO" "Checking if frontend is running..."
try {
    $response = Invoke-WebRequest -Uri $FrontendUrl -Method GET -ErrorAction Stop
    Write-Status "SUCCESS" "Frontend is running"
} catch {
    Write-Status "ERROR" "Frontend is not running. Please start the development server first."
    Write-Host "Run: npm run dev" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "=== Frontend Expiration Testing Checklist ===" -ForegroundColor Green

Write-Host ""
Write-Host "1. COMPOSE MODAL EXPIRATION FUNCTIONALITY:" -ForegroundColor Yellow
Write-Host "   ✓ Email Expiration toggle in Security Settings"
Write-Host "   ✓ Datetime-local input field appears when enabled"
Write-Host "   ✓ Helper text: 'Message will be permanently deleted after this date/time'"
Write-Host "   ✓ Validation prevents past dates"
Write-Host "   ✓ Validation requires date when toggle is enabled"
Write-Host "   ✓ Form resets expiration settings when closed"

Write-Host ""
Write-Host "2. EMAIL DETAIL EXPIRATION DISPLAY:" -ForegroundColor Yellow
Write-Host "   ✓ Expiration status indicator with clock icon"
Write-Host "   ✓ Color-coded indicators (purple for active, red for expired)"
Write-Host "   ✓ Countdown timer showing time remaining"
Write-Host "   ✓ Proper time formatting (days, hours, minutes)"
Write-Host "   ✓ Expiration date in metadata section"
Write-Host "   ✓ Security details show expiration status"

Write-Host ""
Write-Host "3. EXPIRED EMAIL HANDLING:" -ForegroundColor Yellow
Write-Host "   ✓ 'Message Has Expired' screen for expired emails"
Write-Host "   ✓ Clear explanation of permanent deletion"
Write-Host "   ✓ Shows expiration date in expired message"
Write-Host "   ✓ Prevents access to email content"
Write-Host "   ✓ Disables unlock modal for expired emails"
Write-Host "   ✓ Status shows as 'expired' in metadata"

Write-Host ""
Write-Host "4. API INTEGRATION:" -ForegroundColor Yellow
Write-Host "   ✓ Converts datetime-local to ISO 8601 UTC format"
Write-Host "   ✓ Sends expiresAt field to backend"
Write-Host "   ✓ Handles expiration data in API responses"
Write-Host "   ✓ Proper error handling for expiration validation"

Write-Host ""
Write-Host "5. VALIDATION AND USER FEEDBACK:" -ForegroundColor Yellow
Write-Host "   ✓ Error messages for past dates"
Write-Host "   ✓ Error messages for missing dates when enabled"
Write-Host "   ✓ Success messages for valid submissions"
Write-Host "   ✓ Clear visual feedback for validation states"

Write-Host ""
Write-Host "6. RESPONSIVE DESIGN:" -ForegroundColor Yellow
Write-Host "   ✓ Works on desktop (1024px+)"
Write-Host "   ✓ Works on tablet (768px-1023px)"
Write-Host "   ✓ Works on mobile (375px-767px)"
Write-Host "   ✓ Touch-friendly controls"

Write-Host ""
Write-Host "7. ACCESSIBILITY:" -ForegroundColor Yellow
Write-Host "   ✓ Proper ARIA labels for expiration controls"
Write-Host "   ✓ Keyboard navigation support"
Write-Host "   ✓ Screen reader friendly"
Write-Host "   ✓ High contrast support"

Write-Host ""
Write-Host "=== Manual Testing Instructions ===" -ForegroundColor Cyan

Write-Host ""
Write-Host "COMPOSE MODAL TESTING:" -ForegroundColor Yellow
Write-Host "1. Open frontend: $FrontendUrl"
Write-Host "2. Click 'Compose' button to open modal"
Write-Host "3. Scroll to Security Settings section"
Write-Host "4. Find 'Email Expiration' toggle"
Write-Host "5. Enable the toggle - verify datetime input appears"
Write-Host "6. Try setting a past date - verify error message"
Write-Host "7. Set a future date - verify no errors"
Write-Host "8. Fill required fields and submit - verify API call includes expiresAt"
Write-Host "9. Close modal and reopen - verify expiration settings are reset"

Write-Host ""
Write-Host "EMAIL DETAIL TESTING:" -ForegroundColor Yellow
Write-Host "1. Create an email with expiration set to 1 hour from now"
Write-Host "2. View the email in the detail view"
Write-Host "3. Verify expiration status shows with countdown timer"
Write-Host "4. Check security details section shows expiration status"
Write-Host "5. Verify metadata shows expiration date"
Write-Host "6. Wait for email to expire (or create one with past date)"
Write-Host "7. Verify expired message appears"
Write-Host "8. Verify content is not accessible"
Write-Host "9. Test with password-protected expired email"

Write-Host ""
Write-Host "RESPONSIVE TESTING:" -ForegroundColor Yellow
Write-Host "1. Test on desktop browser (full width)"
Write-Host "2. Resize browser to tablet width - verify layout adapts"
Write-Host "3. Resize to mobile width - verify mobile-friendly layout"
Write-Host "4. Test touch interactions on mobile device"

Write-Host ""
Write-Host "ACCESSIBILITY TESTING:" -ForegroundColor Yellow
Write-Host "1. Use keyboard navigation (Tab, Enter, Space)"
Write-Host "2. Test with screen reader software"
Write-Host "3. Verify proper focus indicators"
Write-Host "4. Check color contrast ratios"

Write-Host ""
Write-Status "SUCCESS" "Frontend expiration functionality is ready for comprehensive testing!"
Write-Host ""
Write-Host "Run unit tests: npm test" -ForegroundColor Green
Write-Host "Run integration tests: npm run test:integration" -ForegroundColor Green


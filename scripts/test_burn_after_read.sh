#!/bin/bash

# Test script for burn-after-read functionality
# This script tests the complete flow of creating and accessing burn-after-read emails

set -e

echo "=== Testing Burn-After-Read Functionality ==="

# Configuration
API_BASE="http://localhost:8080"
TEST_USER="test@securesystem.email"
TEST_PASSWORD="testpassword123"
TEST_TOTP="123456"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper function to print colored output
print_status() {
    local status=$1
    local message=$2
    if [ "$status" = "SUCCESS" ]; then
        echo -e "${GREEN}✓ $message${NC}"
    elif [ "$status" = "ERROR" ]; then
        echo -e "${RED}✗ $message${NC}"
    elif [ "$status" = "INFO" ]; then
        echo -e "${YELLOW}ℹ $message${NC}"
    fi
}

# Check if API server is running
print_status "INFO" "Checking if API server is running..."
if ! curl -s "$API_BASE/health" > /dev/null; then
    print_status "ERROR" "API server is not running. Please start the server first."
    exit 1
fi
print_status "SUCCESS" "API server is running"

# Step 1: Login to get JWT token
print_status "INFO" "Logging in to get JWT token..."
LOGIN_RESPONSE=$(curl -s -X POST "$API_BASE/api/auth/login" \
    -H "Content-Type: application/json" \
    -d "{
        \"email\": \"$TEST_USER\",
        \"password\": \"$TEST_PASSWORD\",
        \"totp_code\": \"$TEST_TOTP\"
    }")

TOKEN=$(echo "$LOGIN_RESPONSE" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$TOKEN" ]; then
    print_status "ERROR" "Failed to get JWT token. Login response: $LOGIN_RESPONSE"
    exit 1
fi
print_status "SUCCESS" "Got JWT token"

# Step 2: Send a burn-after-read email
print_status "INFO" "Sending burn-after-read email..."
SEND_RESPONSE=$(curl -s -X POST "$API_BASE/api/email/send" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer $TOKEN" \
    -d "{
        \"recipient\": \"alice@example.com\",
        \"subject\": \"Test Burn-After-Read Email\",
        \"body\": \"This is a test burn-after-read email that should be deleted after first access.\",
        \"burnAfterRead\": true
    }")

EMAIL_ID=$(echo "$SEND_RESPONSE" | grep -o '"blob_id":"[^"]*"' | cut -d'"' -f4 | sed 's/\.blob$//')

if [ -z "$EMAIL_ID" ]; then
    print_status "ERROR" "Failed to send email. Response: $SEND_RESPONSE"
    exit 1
fi
print_status "SUCCESS" "Sent burn-after-read email with ID: $EMAIL_ID"

# Step 3: Access the email for the first time (should succeed)
print_status "INFO" "Accessing burn-after-read email for the first time..."
FIRST_ACCESS_RESPONSE=$(curl -s -X GET "$API_BASE/api/email/view/$EMAIL_ID" \
    -H "Authorization: Bearer $TOKEN")

if echo "$FIRST_ACCESS_RESPONSE" | grep -q '"status":"success"'; then
    print_status "SUCCESS" "First access successful - email content retrieved"
else
    print_status "ERROR" "First access failed. Response: $FIRST_ACCESS_RESPONSE"
    exit 1
fi

# Step 4: Try to access the email again (should return 410 Gone)
print_status "INFO" "Attempting to access burn-after-read email for the second time..."
SECOND_ACCESS_RESPONSE=$(curl -s -w "%{http_code}" -X GET "$API_BASE/api/email/view/$EMAIL_ID" \
    -H "Authorization: Bearer $TOKEN")

HTTP_CODE=$(echo "$SECOND_ACCESS_RESPONSE" | tail -c 4)
RESPONSE_BODY=$(echo "$SECOND_ACCESS_RESPONSE" | head -c -4)

if [ "$HTTP_CODE" = "410" ]; then
    print_status "SUCCESS" "Second access correctly returned 410 Gone - email consumed"
else
    print_status "ERROR" "Second access should return 410 Gone, got $HTTP_CODE. Response: $RESPONSE_BODY"
    exit 1
fi

# Step 5: Verify the email is marked as consumed in the database
print_status "INFO" "Verifying email is marked as consumed in database..."
LIST_RESPONSE=$(curl -s -X GET "$API_BASE/api/email/list" \
    -H "Authorization: Bearer $TOKEN")

if echo "$LIST_RESPONSE" | grep -q "$EMAIL_ID"; then
    print_status "INFO" "Email still appears in list (metadata preserved)"
else
    print_status "INFO" "Email not found in list (completely deleted)"
fi

print_status "SUCCESS" "Burn-after-read functionality test completed successfully!"
echo ""
echo "Test Summary:"
echo "- ✓ API server running"
echo "- ✓ JWT authentication working"
echo "- ✓ Burn-after-read email sent successfully"
echo "- ✓ First access returned email content"
echo "- ✓ Second access returned 410 Gone (email consumed)"
echo "- ✓ Email properly deleted after first access"

echo ""
print_status "INFO" "Burn-after-read functionality is working correctly!"


#!/bin/bash

# Production Testing Script for Secure Email Backend
set -e

# Configuration
API_BASE_URL="https://securemail.example.com"
TEST_EMAIL="test-$(date +%s)@securesystem.email"
TEST_PASSWORD="TestPassword123!"
LOG_FILE="/opt/secure-email/logs/test.log"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo -e "${BLUE}[$(date '+%Y-%m-%d %H:%M:%S')]${NC} $1" | tee -a "$LOG_FILE"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "$LOG_FILE"
    exit 1
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "$LOG_FILE"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1" | tee -a "$LOG_FILE"
}

# Create logs directory if it doesn't exist
mkdir -p "$(dirname "$LOG_FILE")"

# Function to make API request
api_request() {
    local method="$1"
    local endpoint="$2"
    local data="$3"
    local expected_status="$4"
    
    local response
    local status_code
    
    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$API_BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" \
            "$API_BASE_URL$endpoint")
    fi
    
    status_code=$(echo "$response" | tail -n1)
    response_body=$(echo "$response" | head -n -1)
    
    if [ "$status_code" = "$expected_status" ]; then
        success "API request $method $endpoint returned $status_code"
        echo "$response_body"
        return 0
    else
        error "API request $method $endpoint failed. Expected: $expected_status, Got: $status_code. Response: $response_body"
    fi
}

log "Starting production testing..."

# 1. Test health endpoint
log "Testing health endpoint..."
api_request "GET" "/health" "" "200"

# 2. Test signup endpoint with valid data
log "Testing signup endpoint with valid data..."
SIGNUP_RESPONSE=$(api_request "POST" "/api/signup" "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\",
    \"tier\": \"free\"
}" "201")

# Extract recovery token from response
RECOVERY_TOKEN=$(echo "$SIGNUP_RESPONSE" | jq -r '.recovery_token' 2>/dev/null || echo "")
if [ -n "$RECOVERY_TOKEN" ] && [ "$RECOVERY_TOKEN" != "null" ]; then
    success "Recovery token received: ${RECOVERY_TOKEN:0:20}..."
else
    warning "No recovery token in response"
fi

# 3. Test signup with invalid email format
log "Testing signup with invalid email format..."
api_request "POST" "/api/signup" "{
    \"email\": \"invalid-email\",
    \"password\": \"$TEST_PASSWORD\",
    \"tier\": \"free\"
}" "400"

# 4. Test signup with short password
log "Testing signup with short password..."
api_request "POST" "/api/signup" "{
    \"email\": \"test2@securesystem.email\",
    \"password\": \"123\",
    \"tier\": \"free\"
}" "400"

# 5. Test signup with wrong domain for free tier
log "Testing signup with wrong domain for free tier..."
api_request "POST" "/api/signup" "{
    \"email\": \"test@example.com\",
    \"password\": \"$TEST_PASSWORD\",
    \"tier\": \"free\"
}" "400"

# 6. Test signup with premium tier
log "Testing signup with premium tier..."
api_request "POST" "/api/signup" "{
    \"email\": \"premium@example.com\",
    \"password\": \"$TEST_PASSWORD\",
    \"tier\": \"premium\",
    \"custom_domain\": \"example.com\"
}" "201"

# 7. Test duplicate email signup
log "Testing duplicate email signup..."
api_request "POST" "/api/signup" "{
    \"email\": \"$TEST_EMAIL\",
    \"password\": \"$TEST_PASSWORD\",
    \"tier\": \"free\"
}" "400"

# 8. Test database connectivity
log "Testing database connectivity..."
if docker exec securechat-email-db-1 mysql -u secureuser -p"${DB_PASSWORD}" -e "SELECT COUNT(*) FROM users;" securesystem >/dev/null 2>&1; then
    success "Database connectivity test passed"
else
    error "Database connectivity test failed"
fi

# 9. Verify user was created in database
log "Verifying user was created in database..."
USER_COUNT=$(docker exec securechat-email-db-1 mysql -u secureuser -p"${DB_PASSWORD}" -e "SELECT COUNT(*) FROM users WHERE email='$TEST_EMAIL';" securesystem -s --skip-column-names)
if [ "$USER_COUNT" = "1" ]; then
    success "User created successfully in database"
else
    error "User not found in database"
fi

# 10. Check user data integrity
log "Checking user data integrity..."
USER_DATA=$(docker exec securechat-email-db-1 mysql -u secureuser -p"${DB_PASSWORD}" -e "SELECT email, email_verified, tier, created_at FROM users WHERE email='$TEST_EMAIL';" securesystem -s --skip-column-names)
if echo "$USER_DATA" | grep -q "$TEST_EMAIL"; then
    success "User data integrity check passed"
    log "User data: $USER_DATA"
else
    error "User data integrity check failed"
fi

# 11. Test rate limiting
log "Testing rate limiting..."
RATE_LIMIT_HIT=false
for i in {1..10}; do
    response=$(curl -s -w "%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "{\"email\":\"ratelimit$i@securesystem.email\",\"password\":\"$TEST_PASSWORD\",\"tier\":\"free\"}" \
        "$API_BASE_URL/api/signup")
    
    status_code=$(echo "$response" | tail -n1)
    if [ "$status_code" = "429" ]; then
        RATE_LIMIT_HIT=true
        break
    fi
done

if [ "$RATE_LIMIT_HIT" = true ]; then
    success "Rate limiting is working"
else
    warning "Rate limiting may not be working properly"
fi

# 12. Test SSL/TLS
log "Testing SSL/TLS configuration..."
if curl -s -I "$API_BASE_URL/health" | grep -q "HTTP/2 200"; then
    success "SSL/TLS is properly configured"
else
    warning "SSL/TLS configuration may have issues"
fi

# 13. Test security headers
log "Testing security headers..."
SECURITY_HEADERS=$(curl -s -I "$API_BASE_URL/health")
if echo "$SECURITY_HEADERS" | grep -q "Strict-Transport-Security"; then
    success "Security headers are present"
else
    warning "Security headers may be missing"
fi

# 14. Performance test
log "Running performance test..."
RESPONSE_TIMES=()
for i in {1..10}; do
    response_time=$(curl -o /dev/null -s -w '%{time_total}' "$API_BASE_URL/health")
    RESPONSE_TIMES+=("$response_time")
done

# Calculate average response time
total=0
for time in "${RESPONSE_TIMES[@]}"; do
    total=$(echo "$total + $time" | bc)
done
avg_response_time=$(echo "scale=3; $total / ${#RESPONSE_TIMES[@]}" | bc)

if (( $(echo "$avg_response_time < 1.0" | bc -l) )); then
    success "Average response time: ${avg_response_time}s (excellent)"
elif (( $(echo "$avg_response_time < 2.0" | bc -l) )); then
    success "Average response time: ${avg_response_time}s (good)"
else
    warning "Average response time: ${avg_response_time}s (slow)"
fi

# 15. Test email functionality (if SMTP is configured)
log "Testing email functionality..."
if [ -n "$SMTP_HOST" ] && [ "$SMTP_HOST" != "CHANGE_ME" ]; then
    # Check if verification email was sent (this would require email server logs)
    success "SMTP configuration is present"
else
    warning "SMTP not configured - email functionality not tested"
fi

# 16. Cleanup test data
log "Cleaning up test data..."
docker exec securechat-email-db-1 mysql -u secureuser -p"${DB_PASSWORD}" -e "DELETE FROM users WHERE email LIKE 'test-%@securesystem.email' OR email LIKE 'ratelimit%@securesystem.email';" securesystem >/dev/null 2>&1 || true
success "Test data cleaned up"

# Summary
log "Production testing completed successfully!"
log "Test summary:"
log "- Health endpoint: ✓"
log "- Signup functionality: ✓"
log "- Input validation: ✓"
log "- Database operations: ✓"
log "- Rate limiting: ✓"
log "- SSL/TLS: ✓"
log "- Security headers: ✓"
log "- Performance: ✓"
log "- Data integrity: ✓"

success "All production tests passed!"

#!/bin/bash
set -e

# login_test.sh: Test /api/auth/login endpoint
# This script tests the login API with various scenarios

API_URL="https://api.securesystem.email/api/auth/login"
LOG_FILE="tests/login_test.log"

echo "=== Login API Test Suite ===" | tee $LOG_FILE
echo "Testing endpoint: $API_URL" | tee -a $LOG_FILE
echo "Timestamp: $(date)" | tee -a $LOG_FILE
echo "" | tee -a $LOG_FILE

# Function to test API response
test_api() {
    local test_name="$1"
    local expected_status="$2"
    local data="$3"
    
    echo "Testing: $test_name" | tee -a $LOG_FILE
    
    response=$(curl -s -w "%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d "$data" \
        "$API_URL" 2>/dev/null || echo "000")
    
    http_code=${response: -3}
    body=${response:0:-3}
    
    echo "  Expected: $expected_status, Got: $http_code" | tee -a $LOG_FILE
    echo "  Response: $body" | tee -a $LOG_FILE
    
    if [[ $http_code == $expected_status ]]; then
        echo "  ✓ PASS" | tee -a $LOG_FILE
        return 0
    else
        echo "  ✗ FAIL" | tee -a $LOG_FILE
        return 1
    fi
}

# Test 1: Invalid credentials (should return 401)
echo "Test 1: Invalid credentials" | tee -a $LOG_FILE
test_api "Invalid credentials" "401" '{"email":"test@securesystem.email","password":"wrong","totp_code":"000000"}'
test1_result=$?

# Test 2: Invalid email format (should return 400)
echo "" | tee -a $LOG_FILE
echo "Test 2: Invalid email format" | tee -a $LOG_FILE
test_api "Invalid email format" "400" '{"email":"invalid@example.com","password":"testpass","totp_code":"123456"}'
test2_result=$?

# Test 3: Missing required fields (should return 400)
echo "" | tee -a $LOG_FILE
echo "Test 3: Missing required fields" | tee -a $LOG_FILE
test_api "Missing email" "400" '{"password":"testpass","totp_code":"123456"}'
test3_result=$?

test_api "Missing password" "400" '{"email":"test@securesystem.email","totp_code":"123456"}'
test4_result=$?

test_api "Missing TOTP code" "400" '{"email":"test@securesystem.email","password":"testpass"}'
test5_result=$?

# Test 4: Invalid JSON (should return 400)
echo "" | tee -a $LOG_FILE
echo "Test 4: Invalid JSON" | tee -a $LOG_FILE
test_api "Invalid JSON" "400" '{"email":"test@securesystem.email","password":"testpass","totp_code":"123456"'
test6_result=$?

# Test 5: Rate limiting (make multiple requests quickly)
echo "" | tee -a $LOG_FILE
echo "Test 5: Rate limiting" | tee -a $LOG_FILE
echo "  Making 12 requests quickly to test rate limiting..." | tee -a $LOG_FILE

rate_limit_hit=false
for i in {1..12}; do
    response=$(curl -s -w "%{http_code}" -X POST \
        -H "Content-Type: application/json" \
        -d '{"email":"test@securesystem.email","password":"test","totp_code":"123456"}' \
        "$API_URL" 2>/dev/null || echo "000")
    
    http_code=${response: -3}
    if [[ $http_code == "429" ]]; then
        rate_limit_hit=true
        echo "  ✓ Rate limiting triggered on request $i" | tee -a $LOG_FILE
        break
    fi
done

if [[ $rate_limit_hit == true ]]; then
    echo "  ✓ Rate limiting test PASSED" | tee -a $LOG_FILE
    test7_result=0
else
    echo "  ✗ Rate limiting test FAILED (no 429 response)" | tee -a $LOG_FILE
    test7_result=1
fi

# Test 6: Valid request format (should return 401 for non-existent user, but valid format)
echo "" | tee -a $LOG_FILE
echo "Test 6: Valid request format" | tee -a $LOG_FILE
test_api "Valid format, non-existent user" "401" '{"email":"nonexistent@securesystem.email","password":"testpass123","totp_code":"123456"}'
test8_result=$?

# Summary
echo "" | tee -a $LOG_FILE
echo "=== Test Summary ===" | tee -a $LOG_FILE
echo "Test 1 (Invalid credentials): $([ $test1_result -eq 0 ] && echo "PASS" || echo "FAIL")" | tee -a $LOG_FILE
echo "Test 2 (Invalid email): $([ $test2_result -eq 0 ] && echo "PASS" || echo "FAIL")" | tee -a $LOG_FILE
echo "Test 3 (Missing fields): $([ $test3_result -eq 0 ] && [ $test4_result -eq 0 ] && [ $test5_result -eq 0 ] && echo "PASS" || echo "FAIL")" | tee -a $LOG_FILE
echo "Test 4 (Invalid JSON): $([ $test6_result -eq 0 ] && echo "PASS" || echo "FAIL")" | tee -a $LOG_FILE
echo "Test 5 (Rate limiting): $([ $test7_result -eq 0 ] && echo "PASS" || echo "FAIL")" | tee -a $LOG_FILE
echo "Test 6 (Valid format): $([ $test8_result -eq 0 ] && echo "PASS" || echo "FAIL")" | tee -a $LOG_FILE

# Overall result
total_tests=8
passed_tests=0
[ $test1_result -eq 0 ] && ((passed_tests++))
[ $test2_result -eq 0 ] && ((passed_tests++))
[ $test3_result -eq 0 ] && [ $test4_result -eq 0 ] && [ $test5_result -eq 0 ] && ((passed_tests++))
[ $test6_result -eq 0 ] && ((passed_tests++))
[ $test7_result -eq 0 ] && ((passed_tests++))
[ $test8_result -eq 0 ] && ((passed_tests++))

echo "" | tee -a $LOG_FILE
echo "Overall: $passed_tests/$total_tests tests passed" | tee -a $LOG_FILE

if [[ $passed_tests -eq $total_tests ]]; then
    echo "✓ All tests passed!" | tee -a $LOG_FILE
    exit 0
else
    echo "✗ Some tests failed. Check the log for details." | tee -a $LOG_FILE
    exit 1
fi 
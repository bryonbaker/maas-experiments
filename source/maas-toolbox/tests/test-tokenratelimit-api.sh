#!/bin/bash

# Test script for MaaS Toolbox TokenRateLimitPolicy API
# Tests all CRUD operations for token rate limits
#
# Usage:
#   ./test-tokenratelimit-api.sh [MAAS-TOOLBOX_ROUTE_URL]
#   BASE_URL=https://maas-toolbox-maas-toolbox.apps.ocp.domain.com ./test-tokenratelimit-api.sh

############################################################################
# This source file includes portions generated or suggested by
# artificial intelligence tools and subsequently reviewed,
# modified, and validated by human contributors.
#
# Human authorship, design decisions, and final responsibility
# for this code remain with the project contributors.
############################################################################

# Get base URL from command line argument or environment variable, default to localhost
if [ -n "$1" ]; then
    BASE_URL="$1"
elif [ -n "$BASE_URL" ]; then
    BASE_URL="$BASE_URL"
else
    BASE_URL="http://localhost:8080"
fi

# Remove trailing slash if present
BASE_URL="${BASE_URL%/}"

API_BASE="${BASE_URL}/api/v1"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
PASSED=0
FAILED=0
TOTAL=0

# Test data
TEST_TRL_NAME="test-token-rate-limit"
TEST_TRL_NAME_2="test-token-rate-limit-2"

# Function to print test header
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

# Function to run a test and check result
run_test() {
    local test_name="$1"
    local expected_status="$2"
    local method="$3"
    local endpoint="$4"
    local data="$5"
    
    TOTAL=$((TOTAL + 1))
    
    echo -n "Testing: $test_name ... "
    
    if [ -n "$data" ]; then
        response=$(curl -s -k -w "\n%{http_code}" -X "$method" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "${API_BASE}${endpoint}")
    else
        response=$(curl -s -k -w "\n%{http_code}" -X "$method" \
            "${API_BASE}${endpoint}")
    fi
    
    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$http_code" -eq "$expected_status" ]; then
        echo -e "${GREEN}PASS${NC} (HTTP $http_code)"
        PASSED=$((PASSED + 1))
        if [ -n "$body" ] && [ "$body" != "null" ]; then
            echo "$body" | jq . 2>/dev/null || echo "$body"
        fi
        return 0
    else
        echo -e "${RED}FAIL${NC} (Expected HTTP $expected_status, got $http_code)"
        FAILED=$((FAILED + 1))
        echo "Response body:"
        echo "$body" | jq . 2>/dev/null || echo "$body"
        return 1
    fi
}

# Function to cleanup test data
cleanup() {
    echo -e "\n${YELLOW}Cleaning up test data...${NC}"
    curl -s -k -X DELETE "${API_BASE}/tokenratelimits/${TEST_TRL_NAME}" > /dev/null 2>&1
    curl -s -k -X DELETE "${API_BASE}/tokenratelimits/${TEST_TRL_NAME_2}" > /dev/null 2>&1
    echo "Cleanup complete."
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}MaaS Toolbox TokenRateLimitPolicy API Tests${NC}"
echo -e "${BLUE}API Base URL: $API_BASE${NC}"
echo -e "${BLUE}========================================${NC}"

# Test 1: Health check
print_header "Test 1: Health Check"
run_test "Health check" 200 "GET" "/health" ""

# Test 2: Get all token rate limits (initial state)
print_header "Test 2: Get All Token Rate Limits (Initial)"
run_test "Get all token rate limits" 200 "GET" "/tokenratelimits" ""

# Test 3: Create a new token rate limit
print_header "Test 3: Create Token Rate Limit"
run_test "Create token rate limit" 201 "POST" "/tokenratelimits" '{
  "name": "'"$TEST_TRL_NAME"'",
  "limit": 100,
  "window": "1m",
  "tier": "free"
}'

# Test 4: Try to create duplicate (should fail)
print_header "Test 4: Create Duplicate Token Rate Limit (Should Fail)"
run_test "Create duplicate token rate limit" 409 "POST" "/tokenratelimits" '{
  "name": "'"$TEST_TRL_NAME"'",
  "limit": 200,
  "window": "2m",
  "tier": "free"
}'

# Test 5: Get specific token rate limit
print_header "Test 5: Get Specific Token Rate Limit"
run_test "Get token rate limit by name" 200 "GET" "/tokenratelimits/${TEST_TRL_NAME}" ""

# Test 6: Get all token rate limits (should include our test limit)
print_header "Test 6: Get All Token Rate Limits (After Creation)"
run_test "Get all token rate limits" 200 "GET" "/tokenratelimits" ""

# Test 7: Update token rate limit (only limit and window)
print_header "Test 7: Update Token Rate Limit"
run_test "Update token rate limit" 200 "PUT" "/tokenratelimits/${TEST_TRL_NAME}" '{
  "limit": 500,
  "window": "5m"
}'

# Test 8: Verify update by getting the limit
print_header "Test 8: Verify Token Rate Limit Update"
run_test "Get updated token rate limit" 200 "GET" "/tokenratelimits/${TEST_TRL_NAME}" ""

# Test 9: Create another token rate limit
print_header "Test 9: Create Second Token Rate Limit"
run_test "Create second token rate limit" 201 "POST" "/tokenratelimits" '{
  "name": "'"$TEST_TRL_NAME_2"'",
  "limit": 1000,
  "window": "10m",
  "tier": "enterprise"
}'

# Test 10: Delete token rate limit
print_header "Test 10: Delete Token Rate Limit"
run_test "Delete token rate limit" 204 "DELETE" "/tokenratelimits/${TEST_TRL_NAME}" ""

# Test 11: Verify deletion (should return 404)
print_header "Test 11: Verify Token Rate Limit Deletion"
run_test "Get deleted token rate limit" 404 "GET" "/tokenratelimits/${TEST_TRL_NAME}" ""

# Test 12: Delete non-existent limit (should fail)
print_header "Test 12: Delete Non-Existent Token Rate Limit (Should Fail)"
run_test "Delete non-existent token rate limit" 404 "DELETE" "/tokenratelimits/non-existent" ""

# Test 13: Create limit with invalid tier (should fail)
print_header "Test 13: Create Token Rate Limit with Invalid Tier (Should Fail)"
run_test "Create with invalid tier" 500 "POST" "/tokenratelimits" '{
  "name": "invalid-tier-limit",
  "limit": 100,
  "window": "1m",
  "tier": "non-existent-tier"
}'

# Test 14: Create limit with invalid window format (should fail)
print_header "Test 14: Create Token Rate Limit with Invalid Window (Should Fail)"
run_test "Create with invalid window" 400 "POST" "/tokenratelimits" '{
  "name": "invalid-window-limit",
  "limit": 100,
  "window": "invalid",
  "tier": "free"
}'

# Test 15: Create limit with missing required fields (should fail)
print_header "Test 15: Create Token Rate Limit with Missing Fields (Should Fail)"
run_test "Create with missing fields" 400 "POST" "/tokenratelimits" '{
  "name": "incomplete-limit",
  "limit": 100
}'

# Test 16: Clean up remaining test data
print_header "Test 16: Final Cleanup"
run_test "Delete second token rate limit" 204 "DELETE" "/tokenratelimits/${TEST_TRL_NAME_2}" ""

# Print summary
echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}Test Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo -e "Total Tests: $TOTAL"
echo -e "${GREEN}Passed: $PASSED${NC}"
echo -e "${RED}Failed: $FAILED${NC}"

if [ $FAILED -eq 0 ]; then
    echo -e "\n${GREEN}All tests passed!${NC}"
    exit 0
else
    echo -e "\n${RED}Some tests failed!${NC}"
    exit 1
fi

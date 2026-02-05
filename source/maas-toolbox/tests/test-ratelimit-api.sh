#!/bin/bash

# Test script for MaaS Toolbox RateLimitPolicy API
# Tests all CRUD operations for rate limits
#
# Usage:
#   ./test-ratelimit-api.sh [MAAS-TOOLBOX_ROUTE_URL]
#   BASE_URL=https://maas-toolbox-maas-toolbox.apps.ocp.domain.com ./test-ratelimit-api.sh

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
TEST_RL_NAME="test-rate-limit"
TEST_RL_NAME_2="test-rate-limit-2"

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
    curl -s -k -X DELETE "${API_BASE}/ratelimits/${TEST_RL_NAME}" > /dev/null 2>&1
    curl -s -k -X DELETE "${API_BASE}/ratelimits/${TEST_RL_NAME_2}" > /dev/null 2>&1
    echo "Cleanup complete."
}

# Trap to ensure cleanup on exit
trap cleanup EXIT

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}MaaS Toolbox RateLimitPolicy API Tests${NC}"
echo -e "${BLUE}API Base URL: $API_BASE${NC}"
echo -e "${BLUE}========================================${NC}"

# Test 1: Health check
print_header "Test 1: Health Check"
run_test "Health check" 200 "GET" "/health" ""

# Test 2: Get all rate limits (initial state)
print_header "Test 2: Get All Rate Limits (Initial)"
run_test "Get all rate limits" 200 "GET" "/ratelimits" ""

# Test 3: Create a new rate limit
print_header "Test 3: Create Rate Limit"
run_test "Create rate limit" 201 "POST" "/ratelimits" '{
  "name": "'"$TEST_RL_NAME"'",
  "limit": 5,
  "window": "2m",
  "tier": "free"
}'

# Test 4: Try to create duplicate (should fail)
print_header "Test 4: Create Duplicate Rate Limit (Should Fail)"
run_test "Create duplicate rate limit" 409 "POST" "/ratelimits" '{
  "name": "'"$TEST_RL_NAME"'",
  "limit": 10,
  "window": "3m",
  "tier": "free"
}'

# Test 5: Get specific rate limit
print_header "Test 5: Get Specific Rate Limit"
run_test "Get rate limit by name" 200 "GET" "/ratelimits/${TEST_RL_NAME}" ""

# Test 6: Get all rate limits (should include our test limit)
print_header "Test 6: Get All Rate Limits (After Creation)"
run_test "Get all rate limits" 200 "GET" "/ratelimits" ""

# Test 7: Update rate limit (only limit and window)
print_header "Test 7: Update Rate Limit"
run_test "Update rate limit" 200 "PUT" "/ratelimits/${TEST_RL_NAME}" '{
  "limit": 20,
  "window": "5m"
}'

# Test 8: Verify update by getting the limit
print_header "Test 8: Verify Rate Limit Update"
run_test "Get updated rate limit" 200 "GET" "/ratelimits/${TEST_RL_NAME}" ""

# Test 9: Create another rate limit
print_header "Test 9: Create Second Rate Limit"
run_test "Create second rate limit" 201 "POST" "/ratelimits" '{
  "name": "'"$TEST_RL_NAME_2"'",
  "limit": 50,
  "window": "10m",
  "tier": "enterprise"
}'

# Test 10: Delete rate limit
print_header "Test 10: Delete Rate Limit"
run_test "Delete rate limit" 204 "DELETE" "/ratelimits/${TEST_RL_NAME}" ""

# Test 11: Verify deletion (should return 404)
print_header "Test 11: Verify Rate Limit Deletion"
run_test "Get deleted rate limit" 404 "GET" "/ratelimits/${TEST_RL_NAME}" ""

# Test 12: Delete non-existent limit (should fail)
print_header "Test 12: Delete Non-Existent Rate Limit (Should Fail)"
run_test "Delete non-existent rate limit" 404 "DELETE" "/ratelimits/non-existent" ""

# Test 13: Create limit with invalid tier (should fail)
print_header "Test 13: Create Rate Limit with Invalid Tier (Should Fail)"
run_test "Create with invalid tier" 500 "POST" "/ratelimits" '{
  "name": "invalid-tier-limit",
  "limit": 100,
  "window": "1m",
  "tier": "non-existent-tier"
}'

# Test 14: Create limit with invalid window format (should fail)
print_header "Test 14: Create Rate Limit with Invalid Window (Should Fail)"
run_test "Create with invalid window" 400 "POST" "/ratelimits" '{
  "name": "invalid-window-limit",
  "limit": 100,
  "window": "invalid",
  "tier": "free"
}'

# Test 15: Create limit with missing required fields (should fail)
print_header "Test 15: Create Rate Limit with Missing Fields (Should Fail)"
run_test "Create with missing fields" 400 "POST" "/ratelimits" '{
  "name": "incomplete-limit",
  "limit": 100
}'

# Test 16: Clean up remaining test data
print_header "Test 16: Final Cleanup"
run_test "Delete second rate limit" 204 "DELETE" "/ratelimits/${TEST_RL_NAME_2}" ""

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

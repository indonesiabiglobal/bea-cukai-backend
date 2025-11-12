#!/bin/bash

# API Testing Script for Sync Endpoints
# This script tests all sync API endpoints

# Configuration
API_URL="http://localhost:8000"
TOKEN=""  # Add your JWT token here

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║  Bea Cukai Sync API Testing Script    ║${NC}"
echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo ""

# Check if token is set
if [ -z "$TOKEN" ]; then
    echo -e "${RED}ERROR: Please set TOKEN variable in this script${NC}"
    echo -e "${YELLOW}Get token by logging in first:${NC}"
    echo "curl -X POST $API_URL/auth/login -H 'Content-Type: application/json' -d '{\"username\":\"your_username\",\"password\":\"your_password\"}'"
    exit 1
fi

# Function to print test header
print_test() {
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}TEST: $1${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

# Test 1: Check Sync Status (should be idle initially)
print_test "1. Check Initial Sync Status"
echo "Request: GET $API_URL/sync/status"
RESPONSE=$(curl -s -X GET "$API_URL/sync/status" \
  -H "Authorization: Bearer $TOKEN")
echo "Response:"
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"

# Test 2: Get Sync Log
print_test "2. Get Sync Log"
echo "Request: GET $API_URL/sync/log"
RESPONSE=$(curl -s -X GET "$API_URL/sync/log" \
  -H "Authorization: Bearer $TOKEN")
echo "Response (truncated):"
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"

# Test 3: Run Sync Script
print_test "3. Run Sync Script"
echo "Request: POST $API_URL/sync/run"
echo -e "${YELLOW}Running sync script...${NC}"
RESPONSE=$(curl -s -X POST "$API_URL/sync/run" \
  -H "Authorization: Bearer $TOKEN")
echo "Response:"
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"

# Test 4: Check Status While Running
print_test "4. Check Status While Running"
sleep 1
echo "Request: GET $API_URL/sync/status"
RESPONSE=$(curl -s -X GET "$API_URL/sync/status" \
  -H "Authorization: Bearer $TOKEN")
echo "Response:"
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"

# Test 5: Try Running Again (should fail - already running)
print_test "5. Try Running Sync Again (Should Fail)"
echo "Request: POST $API_URL/sync/run"
RESPONSE=$(curl -s -X POST "$API_URL/sync/run" \
  -H "Authorization: Bearer $TOKEN")
echo "Response:"
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"

# Test 6: Wait and Check Final Status
print_test "6. Wait for Sync to Complete"
echo -e "${YELLOW}Waiting 8 seconds for sync to complete...${NC}"
sleep 8

echo "Request: GET $API_URL/sync/status"
RESPONSE=$(curl -s -X GET "$API_URL/sync/status" \
  -H "Authorization: Bearer $TOKEN")
echo "Response:"
echo "$RESPONSE" | jq '.' 2>/dev/null || echo "$RESPONSE"

# Test 7: Get Final Log
print_test "7. Get Final Sync Log"
echo "Request: GET $API_URL/sync/log"
RESPONSE=$(curl -s -X GET "$API_URL/sync/log" \
  -H "Authorization: Bearer $TOKEN")
echo "Response (last 20 lines):"
echo "$RESPONSE" | jq -r '.log' 2>/dev/null | tail -20 || echo "$RESPONSE"

# Summary
echo ""
echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║  Testing Completed                     ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "${YELLOW}Check sync_test.log for detailed sync output${NC}"

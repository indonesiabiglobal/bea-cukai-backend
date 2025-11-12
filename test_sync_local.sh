#!/bin/bash

# Test Sync Script for Local Development
# This script simulates database sync for testing purposes

# Configuration
LOCK_FILE="/tmp/sync_fkk_db_test.lock"
LOG_FILE="./sync_test.log"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging function
log() {
    echo "[$(date '+%Y-%m-%d %H:%M:%S')] $1" | tee -a "$LOG_FILE"
}

# Check if already running
if [ -f "$LOCK_FILE" ]; then
    log "ERROR: Sync is already running"
    echo -e "${RED}ERROR: Sync is already running${NC}"
    exit 2
fi

# Create lock file
touch "$LOCK_FILE"
log "Lock file created: $LOCK_FILE"

# Cleanup function
cleanup() {
    rm -f "$LOCK_FILE"
    log "Lock file removed: $LOCK_FILE"
}
trap cleanup EXIT

# Start sync
echo -e "${GREEN}=== Starting Test Database Sync ===${NC}"
log "=== Starting Test Database Sync ==="

# Simulate database operations
echo -e "${YELLOW}Step 1: Checking database connection...${NC}"
log "Step 1: Checking database connection..."
sleep 1
log "✓ Database connection successful"
echo -e "${GREEN}✓ Database connection successful${NC}"

echo -e "${YELLOW}Step 2: Creating backup...${NC}"
log "Step 2: Creating backup..."
sleep 1
log "✓ Backup created successfully"
echo -e "${GREEN}✓ Backup created successfully${NC}"

echo -e "${YELLOW}Step 3: Syncing data...${NC}"
log "Step 3: Syncing data..."
sleep 2
log "✓ Data synced: 1000 records processed"
echo -e "${GREEN}✓ Data synced: 1000 records processed${NC}"

echo -e "${YELLOW}Step 4: Verifying integrity...${NC}"
log "Step 4: Verifying integrity..."
sleep 1
log "✓ Data integrity verified"
echo -e "${GREEN}✓ Data integrity verified${NC}"

echo -e "${YELLOW}Step 5: Optimizing tables...${NC}"
log "Step 5: Optimizing tables..."
sleep 1
log "✓ Tables optimized"
echo -e "${GREEN}✓ Tables optimized${NC}"

# Final message
echo -e "${GREEN}=== Test Sync Completed Successfully ===${NC}"
log "=== Test Sync Completed Successfully ==="
log "Total execution time: ~6 seconds"

exit 0

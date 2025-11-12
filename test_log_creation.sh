#!/bin/bash

# Test script to verify log file creation
WORKDIR="/tmp"
DATESTR="$(date +%Y%m%d_%H%M%S)"
LOG_FILE="${WORKDIR}/test_log_${DATESTR}.log"
LOG_LATEST="${WORKDIR}/test_log_latest.log"

echo "=== Test Log File Creation ==="
echo "User: $(whoami)"
echo "UID: $(id -u)"
echo "PWD: $(pwd)"
echo "WORKDIR: ${WORKDIR}"
echo "LOG_FILE: ${LOG_FILE}"
echo ""

# Create log file
echo "Creating log file..."
touch "$LOG_FILE"
EXIT_CODE=$?

if [ $EXIT_CODE -eq 0 ]; then
    echo "✓ touch successful"
    ls -l "$LOG_FILE"
    
    # Try to write
    echo "Test write" >> "$LOG_FILE"
    if [ $? -eq 0 ]; then
        echo "✓ Write successful"
        cat "$LOG_FILE"
    else
        echo "✗ Write failed"
    fi
    
    # Try chmod
    chmod 666 "$LOG_FILE" 2>/dev/null
    if [ $? -eq 0 ]; then
        echo "✓ chmod successful"
    else
        echo "✗ chmod failed (might be OK if already writable)"
    fi
    
    # Create symlink
    ln -sf "$LOG_FILE" "$LOG_LATEST"
    if [ $? -eq 0 ]; then
        echo "✓ symlink created"
        ls -l "$LOG_LATEST"
    else
        echo "✗ symlink failed"
    fi
    
    echo ""
    echo "Final check:"
    ls -lh /tmp/test_log*
    
else
    echo "✗ touch failed with exit code: $EXIT_CODE"
    echo "Permission issue in /tmp"
fi

echo ""
echo "=== Test Complete ==="

#!/usr/bin/env bash
# ==============================================================================
# Rutils Mobile - Stop Local Web Preview
# Terminates the running local Flutter web preview instance.
# ==============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
PID_FILE="$APP_DIR/.mobile_web.pid"

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "Stopping Rutils Mobile Web Preview..."

STOPPED=false

# 1. Kill via PID file
if [ -f "$PID_FILE" ]; then
    PID=$(cat "$PID_FILE" 2>/dev/null || true)
    if [ -n "$PID" ] && ps -p "$PID" > /dev/null 2>&1; then
        kill -9 "$PID" 2>/dev/null || true
        echo -e "${GREEN}✓ Terminated process from PID file ($PID)${NC}"
        STOPPED=true
    fi
    rm -f "$PID_FILE"
fi

# 2. Kill any processes on preview ports
PORTS=(8087 8085 5000 8086)
for P in "${PORTS[@]}"; do
    PIDS=$(lsof -ti :"$P" 2>/dev/null || fuser "$P/tcp" 2>/dev/null || true)
    if [ -n "$PIDS" ]; then
        for p in $PIDS; do
            kill -9 "$p" 2>/dev/null || true
            echo -e "${GREEN}✓ Freed port $P (Killed PID: $p)${NC}"
            STOPPED=true
        done
    fi
done

# 3. Kill python http.server serving build/web
PYS=$(pgrep -f "http.server.*build/web" 2>/dev/null || true)
if [ -n "$PYS" ]; then
    for p in $PYS; do
        kill -9 "$p" 2>/dev/null || true
        echo -e "${GREEN}✓ Terminated background web server (PID: $p)${NC}"
        STOPPED=true
    done
fi

if [ "$STOPPED" = true ]; then
    echo -e "${GREEN}✓ Rutils Mobile Web Preview has been completely stopped.${NC}"
else
    echo -e "${YELLOW}No running web preview instance found.${NC}"
fi

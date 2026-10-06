#!/usr/bin/env bash
# ==============================================================================
# Rutils Mobile - Local Web Preview Runner
# Starts the Flutter Mobile Application as a responsive Web preview locally.
# ==============================================================================

set -e

PORT=${1:-8087}
MODE=${2:-"local"} # "local" or "prod"
DAEMON=${3:-"false"}
LIVE_MODE=${4:-"false"} # "live" for flutter run, "static" for compiled release web

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
PID_FILE="$APP_DIR/.mobile_web.pid"
LOG_FILE="$APP_DIR/.mobile_web.log"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m'

echo -e "${BLUE}======================================================${NC}"
echo -e "${GREEN}${BOLD} Rutils Mobile — Local Web Preview Runner ${NC}"
echo -e "${BLUE}======================================================${NC}"

# Stop any process already on this port
OLD_PIDS=$(lsof -ti :"$PORT" 2>/dev/null || true)
if [ -n "$OLD_PIDS" ]; then
    echo -e "${YELLOW}Stopping existing process on port $PORT...${NC}"
    for p in $OLD_PIDS; do
        kill -9 "$p" 2>/dev/null || true
    done
    rm -f "$PID_FILE"
    sleep 1
fi

# Determine target backend
if [ "$MODE" = "prod" ]; then
    API_URL="https://utils.api.srilakshmiretail.in"
    echo -e "Target Backend: ${GREEN}Production ($API_URL)${NC}"
else
    API_URL="http://localhost:8080"
    echo -e "Target Backend: ${GREEN}Local Backend ($API_URL)${NC}"
fi

echo -e "Preview Port:   ${CYAN}$PORT${NC}"
cd "$APP_DIR"

if [ "$LIVE_MODE" = "live" ] || [ "$LIVE_MODE" = "debug" ]; then
    echo -e "${CYAN}Starting Flutter interactive Web Server in Live Mode...${NC}"
    if [ "$DAEMON" = "daemon" ] || [ "$DAEMON" = "true" ] || [ "$DAEMON" = "-d" ]; then
        nohup flutter run -d web-server \
            --web-port="$PORT" \
            --web-hostname=0.0.0.0 \
            --dart-define=API_URL="$API_URL" > "$LOG_FILE" 2>&1 &
        echo $! > "$PID_FILE"
        echo -e "${GREEN}✓ Live web server started in background (PID: $(cat "$PID_FILE"))${NC}"
        echo -e "${CYAN}👉 Open in browser: http://localhost:$PORT${NC}"
    else
        flutter run -d web-server \
            --web-port="$PORT" \
            --web-hostname=0.0.0.0 \
            --dart-define=API_URL="$API_URL"
    fi
else
    # Compiled fast Web preview
    if [ ! -f "$APP_DIR/build/web/index.html" ]; then
        echo -e "${YELLOW}Compiled web build not found. Building release web bundle now...${NC}"
        flutter build web --release --dart-define=API_URL="$API_URL"
    fi

    echo -e "${GREEN}✓ Launching optimized Web Preview on port $PORT...${NC}"

    if [ "$DAEMON" = "daemon" ] || [ "$DAEMON" = "true" ] || [ "$DAEMON" = "-d" ]; then
        nohup python3 -m http.server -d "$APP_DIR/build/web" "$PORT" > "$LOG_FILE" 2>&1 &
        SERVER_PID=$!
        disown "$SERVER_PID" 2>/dev/null || true
        echo "$SERVER_PID" > "$PID_FILE"
        sleep 1
        echo -e "${GREEN}✓ Local Web Preview is running in background (PID: $SERVER_PID)${NC}"
        echo -e "${CYAN}${BOLD}👉 Open in browser: http://localhost:$PORT${NC}"
        echo -e "To stop: ./scripts/stop_local_web.sh"
    else
        echo -e "${CYAN}${BOLD}👉 Open in browser: http://localhost:$PORT${NC}"
        echo -e "Press Ctrl+C to stop."
        python3 -m http.server -d "$APP_DIR/build/web" "$PORT"
    fi
fi

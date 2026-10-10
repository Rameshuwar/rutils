#!/usr/bin/env bash

# ==============================================================================
# Local Development Script for File Converter (Backend + Frontend)
# ==============================================================================

set -e

# Terminal colors
CYAN='\033[0;36m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function to clean up background processes on Ctrl+C (SIGINT/SIGTERM)
cleanup() {
    echo -e "\n${YELLOW}[!] Stopping all services...${NC}"
    if [[ -n "$BACKEND_PID" ]]; then
        kill "$BACKEND_PID" 2>/dev/null || true
    fi
    if [[ -n "$FRONTEND_PID" ]]; then
        kill "$FRONTEND_PID" 2>/dev/null || true
    fi
    echo -e "${GREEN}[✓] Stopped background services successfully.${NC}"
    exit 0
}

trap cleanup INT TERM EXIT

echo -e "${CYAN}=================================================="
echo -e " Starting Local Development Environment"
echo -e "==================================================${NC}"

export PATH="$PATH:$HOME/go/bin:$(go env GOPATH 2>/dev/null)/bin"

# 1. Regenerate Swagger Docs
echo -e "${GREEN}[1/3] Generating latest Swagger documentation...${NC}"
if command -v swag >/dev/null 2>&1; then
    swag init -g cmd/server/main.go -o docs
elif [ -x "$HOME/go/bin/swag" ]; then
    "$HOME/go/bin/swag" init -g cmd/server/main.go -o docs
elif [ -n "$(go env GOPATH 2>/dev/null)" ] && [ -x "$(go env GOPATH)/bin/swag" ]; then
    "$(go env GOPATH)/bin/swag" init -g cmd/server/main.go -o docs
else
    go run github.com/swaggo/swag/cmd/swag@v1.16.2 init -g cmd/server/main.go -o docs
fi
echo -e "${GREEN}[✓] Swagger documentation updated.${NC}"

# 2. Start Go Backend
echo -e "${GREEN}[2/3] Starting Go Backend API (port 8080)...${NC}"
go run ./cmd/server/main.go &
BACKEND_PID=$!

# 3. Check and start Frontend
echo -e "${GREEN}[3/3] Preparing Vite Frontend dev server...${NC}"
if [ ! -d "frontend/node_modules" ]; then
    echo -e "${YELLOW}[!] node_modules missing in frontend/. Running 'npm install'...${NC}"
    (cd frontend && npm install)
fi

echo -e "${GREEN}[✓] Starting Vite Frontend dev server...${NC}"
(cd frontend && npm run dev) &
FRONTEND_PID=$!

echo -e "\n${CYAN}=================================================="
echo -e " Services Running:"
echo -e "  Backend API:  http://localhost:8080"
echo -e "  Swagger UI:   http://localhost:8080/swagger/"
echo -e "  Frontend UI:  http://localhost:5173"
echo -e " Press [Ctrl+C] to stop all services"
echo -e "==================================================${NC}\n"

# Wait for background jobs to finish or signal
wait

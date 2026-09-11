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

# 1. Start Go Backend
echo -e "${GREEN}[1/2] Starting Go Backend API (port 8080)...${NC}"
go run ./cmd/server/main.go &
BACKEND_PID=$!

# 2. Check and start Frontend
echo -e "${GREEN}[2/2] Preparing Vite Frontend dev server...${NC}"
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

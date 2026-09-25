#!/bin/bash

# Ensure script executes in persona-tests directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo "========================================================="
echo " Universal Converter - FULL Persona Test Suite"
echo "========================================================="
echo ""

# ----------------------------------------------------------
# Detect whether a backend is already running on :8080
# ----------------------------------------------------------
BACKEND_ALREADY_UP=false
if curl -s -o /dev/null --max-time 1 http://localhost:8080/swagger/doc.json 2>/dev/null; then
    BACKEND_ALREADY_UP=true
    echo "[i] Detected existing backend on :8080 — will reuse it."
fi

# ----------------------------------------------------------
# Start our own backend only if one isn't already up
# ----------------------------------------------------------
OWN_BACKEND_PID=""
if [ "$BACKEND_ALREADY_UP" = false ]; then
    echo "[i] Starting Go API server..."
    cd ../
    go run cmd/server/main.go > /dev/null 2>&1 &
    OWN_BACKEND_PID=$!
    cd "$SCRIPT_DIR"
    echo "[i] Waiting for backend to be ready on :8080..."
    sleep 2
fi

# ----------------------------------------------------------
# Run UI tests (tests/ui/ only)
# ----------------------------------------------------------
echo ""
echo "[1/2] Running Frontend UI Tests..."
npx playwright test tests/ui/
UI_CODE=$?

# ----------------------------------------------------------
# Run API tests (tests/api/ only)
# ----------------------------------------------------------
echo ""
echo "[2/2] Running Backend API Tests..."
npx playwright test tests/api/
API_CODE=$?

# ----------------------------------------------------------
# Cleanup only the backend we started
# ----------------------------------------------------------
if [ -n "$OWN_BACKEND_PID" ]; then
    echo ""
    echo "[i] Shutting down our backend (PID: $OWN_BACKEND_PID)..."
    kill "$OWN_BACKEND_PID" 2>/dev/null || true
fi

# ----------------------------------------------------------
# Summary
# ----------------------------------------------------------
echo ""
echo "========================================================="
if [ $UI_CODE -eq 0 ] && [ $API_CODE -eq 0 ]; then
    echo "🌟 SUCCESS: ALL Persona Tests (UI and API) passed perfectly!"
else
    echo "⚠️  Some tests failed. Please review the output above."
fi
echo "========================================================="
echo "Opening the combined HTML report..."
npx playwright show-report

exit $((UI_CODE + API_CODE))
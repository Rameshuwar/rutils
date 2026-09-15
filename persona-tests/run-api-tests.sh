#!/bin/bash

# Ensure script executes in persona-tests directory where tests reside
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo "========================================================="
echo " Universal Converter - Persona API Test Runner"
echo "========================================================="
echo ""

echo "Starting Go API Server in the background..."
cd ../ # Go to root of project
go run cmd/server/main.go > /dev/null 2>&1 &
SERVER_PID=$!

echo "Waiting for server to be ready on port 8080..."
sleep 2

cd "$SCRIPT_DIR"

echo "Running API Persona Tests using Playwright..."
npx playwright test tests/api/

EXIT_CODE=$?

echo ""
echo "========================================================="
if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ All API Persona tests passed!"
else
    echo "❌ Some API tests failed. Please check the report."
fi
echo "========================================================="

echo "Shutting down Go API Server (PID: $SERVER_PID)..."
kill $SERVER_PID 2>/dev/null || true

echo "Opening the beautifully detailed HTML report..."
npx playwright show-report

exit $EXIT_CODE

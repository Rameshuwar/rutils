#!/bin/bash

# Ensure script executes in persona-tests directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo "========================================================="
echo " Universal Converter - FULL Persona Test Suite"
echo "========================================================="
echo ""

echo "[1/2] Running Frontend UI Tests..."
./run-ui-tests.sh
UI_CODE=$?

echo ""
echo "[2/2] Running Backend API Tests..."
./run-api-tests.sh
API_CODE=$?

echo ""
echo "========================================================="
if [ $UI_CODE -eq 0 ] && [ $API_CODE -eq 0 ]; then
    echo "🌟 SUCCESS: ALL Persona Tests (UI and API) passed perfectly!"
else
    echo "⚠️  Some tests failed. Please review the output above."
fi
echo "========================================================="
echo "Opening the combined HTML report containing ALL detailed steps..."
npx playwright show-report

exit $((UI_CODE + API_CODE))

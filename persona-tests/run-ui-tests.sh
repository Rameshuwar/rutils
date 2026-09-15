#!/bin/bash

# Ensure script executes in persona-tests directory where tests reside
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

echo "========================================================="
echo " Universal Converter - Persona UI Test Runner"
echo "========================================================="
echo ""

echo "Running UI Persona Tests using Playwright..."
npx playwright test "$@"

EXIT_CODE=$?

echo ""
echo "========================================================="
if [ $EXIT_CODE -eq 0 ]; then
    echo "✅ All UI Persona tests passed!"
else
    echo "❌ Some tests failed. Please check the report."
fi
echo "========================================================="
echo "Opening the detailed HTML report..."
npx playwright show-report

exit $EXIT_CODE

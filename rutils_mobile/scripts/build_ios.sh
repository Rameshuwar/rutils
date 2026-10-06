#!/usr/bin/env bash
# ==============================================================================
# Rutils Mobile - iOS Build Script
# Prepares iOS Runner & Release Archive for App Store / TestFlight deployment.
# ==============================================================================

set -e

PROD_API="https://utils.api.srilakshmiretail.in"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m'

echo -e "${BLUE}======================================================${NC}"
echo -e "${GREEN}${BOLD} Rutils Mobile — iOS Build Engine ${NC}"
echo -e "${BLUE}======================================================${NC}"
echo -e "Bundle Identifier:     ${CYAN}com.srilakshmiretail.rutils${NC}"
echo -e "Production API Target: ${CYAN}$PROD_API${NC}"
echo ""

cd "$APP_DIR"

# Check if running on macOS or Linux
OS_NAME="$(uname -s)"
if [ "$OS_NAME" != "Darwin" ]; then
    echo -e "${YELLOW}Notice: Full iOS IPA archiving and code signing require macOS with Xcode.${NC}"
    echo -e "${CYAN}Building platform-agnostic Flutter iOS assets and bundle dependencies...${NC}"
    flutter build bundle
    echo -e "${GREEN}✓ Flutter assets and Dart AOT bundles compiled for iOS target.${NC}"
    echo ""
    echo "To finalize App Store / TestFlight IPA packaging on a Mac:"
    echo "  1. Transfer rutils_mobile to macOS"
    echo "  2. Run: flutter build ipa --release --dart-define=API_URL=$PROD_API"
    exit 0
fi

echo -e "${CYAN}==> Building iOS Release (no-codesign)...${NC}"
flutter build ios --release --no-codesign --dart-define=API_URL="$PROD_API"

echo -e "${GREEN}✓ iOS Runner release build generated:${NC}"
ls -lh "$APP_DIR/build/ios/iphoneos" 2>/dev/null || true
echo ""
echo -e "${CYAN}To generate signed IPA for App Store submission:${NC}"
echo "  flutter build ipa --release --dart-define=API_URL=$PROD_API"

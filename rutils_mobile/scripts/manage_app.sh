#!/usr/bin/env bash
# ==============================================================================
# Rutils Mobile - Interactive Management CLI
# Start, stop, test, run, and build across Android, iOS, and Web.
# ==============================================================================

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

GREEN='\033[0;32m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BOLD='\033[1m'
NC='\033[0m'

clear_screen() {
    clear 2>/dev/null || true
}

show_header() {
    echo -e "${BLUE}=================================================================${NC}"
    echo -e "${GREEN}${BOLD}     🚀 RUTILS MOBILE APPLICATION DASHBOARD & MANAGER           ${NC}"
    echo -e "${BLUE}=================================================================${NC}"
    echo -e " Package:     ${CYAN}com.srilakshmiretail.rutils${NC}"
    echo -e " Backend:     ${CYAN}Local (8080) ⇄ Production (utils.api.srilakshmiretail.in)${NC}"
    echo -e " Path:        $APP_DIR"
    echo -e "${BLUE}-----------------------------------------------------------------${NC}"
}

menu() {
    show_header
    echo -e "${BOLD}Select an action:${NC}"
    echo ""
    echo -e "  ${CYAN}[1]${NC} 🌐 Start Local Web Preview (Browser responsive view on port 8087)"
    echo -e "  ${CYAN}[2]${NC} 🛑 Stop Local Web Preview"
    echo -e "  ${CYAN}[3]${NC} 📱 Run on Connected Android Device / Emulator"
    echo -e "  ${CYAN}[4]${NC} 💻 Run natively on Linux Desktop"
    echo -e "  ${CYAN}[5]${NC} 📦 Build Android Release APK (Direct install / sideload)"
    echo -e "  ${CYAN}[6]${NC} 🚀 Build Google Play Store App Bundle (AAB deployable)"
    echo -e "  ${CYAN}[7]${NC} 🍏 Build iOS Release Bundle / Assets"
    echo -e "  ${CYAN}[8]${NC} 🔍 Run Flutter Doctor & Code Analyzer"
    echo -e "  ${CYAN}[9]${NC} 🧪 Run Automated Tests"
    echo -e "  ${CYAN}[0]${NC} 🚪 Exit"
    echo ""
    read -p "Enter selection [0-9]: " choice

    case "$choice" in
        1)
            echo ""
            read -p "Target backend [1=Local 8080 (default), 2=Production]: " bchoice
            MODE="local"
            if [ "$bchoice" = "2" ]; then
                MODE="prod"
            fi
            read -p "Run in background daemon mode? [y/N]: " dchoice
            DAEMON="false"
            if [ "$dchoice" = "y" ] || [ "$dchoice" = "Y" ]; then
                DAEMON="daemon"
            fi
            "$SCRIPT_DIR/run_local_web.sh" 8087 "$MODE" "$DAEMON"
            read -p "Press Enter to return to menu..."
            menu
            ;;
        2)
            echo ""
            "$SCRIPT_DIR/stop_local_web.sh"
            read -p "Press Enter to return to menu..."
            menu
            ;;
        3)
            echo ""
            echo -e "${CYAN}Running on connected device/emulator...${NC}"
            cd "$APP_DIR" && flutter run
            read -p "Press Enter to return to menu..."
            menu
            ;;
        4)
            echo ""
            echo -e "${CYAN}Running natively on Linux Desktop...${NC}"
            cd "$APP_DIR" && flutter run -d linux
            read -p "Press Enter to return to menu..."
            menu
            ;;
        5)
            echo ""
            "$SCRIPT_DIR/build_android.sh" apk
            read -p "Press Enter to return to menu..."
            menu
            ;;
        6)
            echo ""
            "$SCRIPT_DIR/build_android.sh" aab
            read -p "Press Enter to return to menu..."
            menu
            ;;
        7)
            echo ""
            "$SCRIPT_DIR/build_ios.sh"
            read -p "Press Enter to return to menu..."
            menu
            ;;
        8)
            echo ""
            echo -e "${CYAN}Running Flutter Doctor & Analyze...${NC}"
            cd "$APP_DIR"
            flutter doctor
            echo ""
            flutter analyze
            read -p "Press Enter to return to menu..."
            menu
            ;;
        9)
            echo ""
            echo -e "${CYAN}Running test suite...${NC}"
            cd "$APP_DIR" && flutter test
            read -p "Press Enter to return to menu..."
            menu
            ;;
        0)
            echo -e "${GREEN}Goodbye!${NC}"
            exit 0
            ;;
        *)
            echo -e "${RED}Invalid selection. Please try again.${NC}"
            sleep 1
            menu
            ;;
    esac
}

menu

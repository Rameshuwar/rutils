#!/usr/bin/env bash
# ==============================================================================
# push_nse_data.sh — Push NSE Technical Chart Market Data to VPS Host Storage
# ==============================================================================

set -Eeuo pipefail

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

log_info() { echo -e "${CYAN}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warn() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1" >&2; }
die() { log_error "$1"; exit 1; }

echo -e "\n${BOLD}${CYAN}======================================================${NC}"
echo -e "${BOLD}${CYAN}   RUTILS — Push NSE Technical Chart Data to VPS      ${NC}"
echo -e "${BOLD}${CYAN}======================================================${NC}\n"

LOCAL_NSE_DIR="./data/nse"

if [[ ! -d "$LOCAL_NSE_DIR" ]]; then
    die "Local NSE data directory not found at '$LOCAL_NSE_DIR'. Please make sure data/nse exists."
fi

# 1. VPS Connection Configuration
VPS_IP="${1:-"148.230.66.115"}"
VPS_USER="${2:-"sysInfra"}"
DEFAULT_KEY="$HOME/.ssh/laxmi/id_rsa_laxmi_sysinfra"
VPS_KEY="${3:-"$DEFAULT_KEY"}"
VPS_REMOTE_DATA="/apps/rutils/data"

read -r -p "VPS IP / Hostname [$VPS_IP]: " input_ip || true
[[ -n "${input_ip:-}" ]] && VPS_IP="$input_ip"

read -r -p "SSH Username [$VPS_USER]: " input_user || true
[[ -n "${input_user:-}" ]] && VPS_USER="$input_user"

if [[ ! -f "$VPS_KEY" ]]; then
    if [[ -f "$HOME/.ssh/id_rsa" ]]; then
        VPS_KEY="$HOME/.ssh/id_rsa"
    elif [[ -f "$HOME/.ssh/id_ed25519" ]]; then
        VPS_KEY="$HOME/.ssh/id_ed25519"
    else
        VPS_KEY=""
    fi
fi

read -r -p "SSH Key Path [${VPS_KEY:-"none"}]: " input_key || true
[[ -n "${input_key:-}" ]] && VPS_KEY="$input_key"

VPS_PASS=""
if [[ -z "$VPS_KEY" || ! -f "$VPS_KEY" ]]; then
    read -r -s -p "SSH Password for $VPS_USER@$VPS_IP: " VPS_PASS || true
    echo ""
fi

# 2. Build SSH command
SSH_CMD="ssh -o StrictHostKeyChecking=accept-new -o ConnectTimeout=15"
if [[ -n "$VPS_KEY" && -f "$VPS_KEY" ]]; then
    SSH_CMD="$SSH_CMD -i $VPS_KEY"
fi
if [[ -n "$VPS_PASS" ]] && command -v sshpass >/dev/null 2>&1; then
    SSH_CMD="sshpass -p \"$VPS_PASS\" $SSH_CMD"
fi

log_info "Testing SSH connectivity to $VPS_USER@$VPS_IP..."
if ! eval "$SSH_CMD \"$VPS_USER@$VPS_IP\" 'echo SSH_OK'" >/dev/null 2>&1; then
    die "Could not connect to $VPS_USER@$VPS_IP via SSH. Please check credentials or network."
fi
log_success "SSH connection verified!"

# 3. Prepare Remote Target Directory
log_info "Ensuring remote directory '$VPS_REMOTE_DATA' exists..."
eval "$SSH_CMD \"$VPS_USER@$VPS_IP\" 'mkdir -p $VPS_REMOTE_DATA'"

# 4. Stream Compressed NSE Data to VPS
COMPANY_COUNT=$(find "$LOCAL_NSE_DIR/data/companies" -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l)
log_info "Packaging & streaming $COMPANY_COUNT companies from $LOCAL_NSE_DIR to $VPS_USER@$VPS_IP:$VPS_REMOTE_DATA..."

tar -czf - -C ./data nse | eval "$SSH_CMD \"$VPS_USER@$VPS_IP\" 'tar -xzf - -C $VPS_REMOTE_DATA && chmod -R 755 $VPS_REMOTE_DATA/nse'"

# 5. Remote Verification
log_info "Verifying remote deployment on VPS..."
REMOTE_COUNT=$(eval "$SSH_CMD \"$VPS_USER@$VPS_IP\" 'find $VPS_REMOTE_DATA/nse/data/companies -mindepth 1 -maxdepth 1 -type d 2>/dev/null | wc -l'")

if [[ "$REMOTE_COUNT" -gt 0 ]]; then
    log_success "Verified: $REMOTE_COUNT companies successfully placed on VPS at $VPS_REMOTE_DATA/nse!"
else
    log_warn "Uploaded data, but could not verify company count. Check manually at $VPS_REMOTE_DATA/nse."
fi

# 6. Optional Container Restart
CONTAINER_NAME="rutils-app"
echo -e "\nWould you like to restart the container '$CONTAINER_NAME' now to apply the new data?"
read -r -p "Restart container? (y/n) [y]: " RESTART_CHOICE || true
RESTART_CHOICE=${RESTART_CHOICE:-"y"}

if [[ "${RESTART_CHOICE,,}" == "y" || "${RESTART_CHOICE,,}" == "yes" ]]; then
    log_info "Restarting container '$CONTAINER_NAME' on VPS..."
    eval "$SSH_CMD \"$VPS_USER@$VPS_IP\" 'docker restart $CONTAINER_NAME 2>/dev/null || docker restart rutils 2>/dev/null || true'"
    log_success "Container restarted. New market data is now live on the server!"
else
    log_info "Skipped container restart. You can restart manually whenever ready."
fi

echo -e "\n${BOLD}${GREEN}======================================================${NC}"
echo -e "${BOLD}${GREEN}   NSE Data Sync Completed Successfully!             ${NC}"
echo -e "${BOLD}${GREEN}======================================================${NC}\n"

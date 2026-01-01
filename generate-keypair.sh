#!/bin/bash
set -e

# --- Configuration for "Pretty" Output ---
# ANSI Color Codes
BOLD="\033[1m"
GREEN="\033[0;32m"
RED="\033[0;31m"
BLUE="\033[0;34m"
RESET="\033[0m"

# --- Logic ---

# 1. Create a secure temporary directory
TMP_DIR=$(mktemp -d)

# 2. Set trap to clean up immediately on exit (successful or not)
trap 'rm -rf "$TMP_DIR"' EXIT

# 3. Generate the keys quietly
# -t ed25519: Key type
# -f: Output file path
# -N "": No passphrase
# -q: Quiet mode
# -C: Comment (set to "temp-generated-key")
ssh-keygen -t ed25519 -f "$TMP_DIR/key" -N "" -q -C "temp-generated-key"

# 4. Helper function to normalize base64 output for Arch/Mac compatibility
# Mac 'base64' splits lines by default; Arch 'base64' needs -w0. 
# Using 'tr -d' handles both.
get_b64() {
    cat "$1" | base64 | tr -d '\n'
}

PUB_B64=$(get_b64 "$TMP_DIR/key.pub")
PRIV_B64=$(get_b64 "$TMP_DIR/key")

# --- Output ---

echo ""
echo -e "${BLUE}=== SSH Key Generator (Ed25519) ===${RESET}"
echo ""

echo -e "${BOLD}1. Public Key ${RESET}(${GREEN}Safe to share${RESET})"
echo "----------------------------------------"
echo -e "${GREEN}${PUB_B64}${RESET}"
echo "----------------------------------------"
echo ""

echo -e "${BOLD}2. Private Key ${RESET}(${RED}SECRET! Keep safe${RESET})"
echo "----------------------------------------"
echo -e "${RED}${PRIV_B64}${RESET}"
echo "----------------------------------------"
echo ""

echo -e "${BLUE}=== End ===${RESET}"
echo ""

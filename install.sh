#!/usr/bin/env bash
set -euo pipefail

GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
ok()   { echo -e "${GREEN}✓${NC} $*"; }
warn() { echo -e "${YELLOW}⚠${NC}  $*"; }
fail() { echo -e "${RED}✗${NC} $*"; exit 1; }

echo "=== gh-next installer ==="
echo

# Prerequisites
command -v gh &>/dev/null || fail "gh not found. Install: https://cli.github.com"
ok "gh $(gh --version | head -1 | awk '{print $3}')"

gh auth status &>/dev/null || fail "gh not authenticated. Run: gh auth login"
ok "gh authenticated as $(gh api user --jq '.login')"

command -v jq &>/dev/null || { warn "jq not found — installing via brew..."; brew install jq; }
ok "jq $(jq --version)"

command -v terminal-notifier &>/dev/null || { warn "terminal-notifier not found — installing via brew..."; brew install terminal-notifier; }
ok "terminal-notifier $(terminal-notifier -version 2>/dev/null | head -1)"

# Install / upgrade extension
echo
if gh extension list | grep -q "maastrich/gh-next"; then
    echo "Upgrading gh-next..."
    gh extension upgrade gh-next
    ok "gh-next upgraded"
else
    echo "Installing gh-next..."
    gh extension install maastrich/gh-next
    ok "gh-next installed"
fi

# Schedule (weekdays 8am–6pm, every hour)
echo
echo "Setting up recurring schedule (weekdays 8am–6pm, hourly)..."
gh next program
echo

# First run
echo "Running first fetch..."
gh next status
echo

ok "Done!"
echo
echo "Commands:"
echo "  gh next                   show cached status"
echo "  gh next status            refresh now"
echo "  gh next program --show    show schedule"
echo "  gh next program --remove  remove schedule"

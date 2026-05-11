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

# Bootstrap (deps + launcher script)
echo
gh next bootstrap
echo

# Schedule (weekdays 8am–6pm, every hour)
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

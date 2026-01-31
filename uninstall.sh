#!/bin/bash
set -e

echo "🗑️  Uninstalling kubecheck..."

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Remove kubecheck binaries only
if [ -f /usr/local/bin/kubecheck ]; then
    echo "  Removing CLI binary..."
    sudo rm -f /usr/local/bin/kubecheck
    echo -e "${GREEN}✓${NC} CLI removed"
fi

if [ -d /usr/local/lib/kubecheck ]; then
    echo "  Removing rule engine..."
    sudo rm -rf /usr/local/lib/kubecheck
    echo -e "${GREEN}✓${NC} Rule engine removed"
fi

echo ""
echo -e "${GREEN}✅ kubecheck uninstalled successfully${NC}"
echo ""
echo -e "${YELLOW}⚠️  Note: Prerequisites (Go, GHC, Cabal, Helm) were NOT removed${NC}"
echo ""
echo "These tools may be used by other projects on your system."
echo "To remove them manually (⚠️  DANGEROUS - may break other tools):"
echo ""
echo "  # Remove Go"
echo "  sudo rm -rf /usr/local/go"
echo "  # Remove from PATH in ~/.bashrc or ~/.zshrc"
echo ""
echo "  # Remove Haskell (via ghcup)"
echo "  ghcup nuke"
echo ""
echo "  # Remove Helm"
echo "  sudo rm /usr/local/bin/helm  # if installed via script"
echo "  brew uninstall helm           # if installed via brew"
echo ""
echo -e "${RED}⚠️  Only remove these if you're CERTAIN no other projects use them!${NC}"

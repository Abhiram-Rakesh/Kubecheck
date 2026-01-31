#!/bin/bash
set -e

echo "🗑️  Uninstalling kubecheck..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

# Remove binaries
if [ -f /usr/local/bin/kubecheck ]; then
    echo "  Removing CLI binary..."
    sudo rm -f /usr/local/bin/kubecheck
    echo -e "${GREEN}✓${NC} CLI removed"
else
    echo -e "${YELLOW}⚠${NC}  CLI not found"
fi

# Remove rule engine
if [ -d /usr/local/lib/kubecheck ]; then
    echo "  Removing rule engine..."
    sudo rm -rf /usr/local/lib/kubecheck
    echo -e "${GREEN}✓${NC} Rule engine removed"
else
    echo -e "${YELLOW}⚠${NC}  Rule engine not found"
fi

echo ""
echo -e "${GREEN}✅ Uninstallation complete!${NC}"

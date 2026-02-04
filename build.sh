#!/bin/bash
set -e

echo "🔨 Building kubecheck..."

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

PROJECT_ROOT=$(pwd)
REQUIRED_GO_VERSION="1.21"
GO_INSTALL_DIR="/usr/local/go"
GO_TARBALL="go1.22.1.linux-amd64.tar.gz"
GO_DOWNLOAD_URL="https://go.dev/dl/${GO_TARBALL}"

echo "📋 Checking prerequisites..."

install_go() {
    echo -e "${YELLOW}⚠ Go not found or version too old. Installing Go...${NC}"

    sudo rm -rf "$GO_INSTALL_DIR"
    curl -fsSL "$GO_DOWNLOAD_URL" -o /tmp/$GO_TARBALL

    sudo tar -C /usr/local -xzf /tmp/$GO_TARBALL
    rm /tmp/$GO_TARBALL

    # Ensure Go is in PATH for this script
    export PATH=/usr/local/go/bin:$PATH

    if ! command -v go &>/dev/null; then
        echo -e "${RED}❌ Go installation failed${NC}"
        exit 1
    fi

    echo -e "${GREEN}✓ Go installed successfully${NC}"
}

version_ge() {
    # returns 0 if $1 >= $2
    [ "$(printf '%s\n' "$2" "$1" | sort -V | head -n1)" = "$2" ]
}

if command -v go &>/dev/null; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo -e "${GREEN}✓${NC} Go $GO_VERSION found"

    if ! version_ge "$GO_VERSION" "$REQUIRED_GO_VERSION"; then
        install_go
    fi
else
    install_go
fi

echo ""
echo "🔧 Building Go CLI..."
cd "$PROJECT_ROOT/cmd/kubecheck"

go mod tidy
go build -o kubecheck
echo -e "${GREEN}✓${NC} Go CLI built"

echo ""
echo "📦 Installing to /usr/local..."

sudo mkdir -p /usr/local/bin
sudo mkdir -p /usr/local/lib/kubecheck

echo "  Installing kubecheck CLI..."
sudo cp kubecheck /usr/local/bin/kubecheck
sudo chmod +x /usr/local/bin/kubecheck

echo ""
echo -e "${GREEN}✅ Installation complete!${NC}"
echo ""
echo "Try it out:"
echo "  kubecheck examples/deployment.yaml"
echo ""
echo "To uninstall:"
echo "  ./uninstall.sh"

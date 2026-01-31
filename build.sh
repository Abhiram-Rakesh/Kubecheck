#!/bin/bash
set -e

echo "🔨 Building kubecheck..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check prerequisites
echo "📋 Checking prerequisites..."

# Check Go
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed${NC}"
    echo "Please install Go >= 1.21 from https://go.dev/dl/"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo -e "${GREEN}✓${NC} Go $GO_VERSION found"

# Check GHC (via ghcup)
if ! command -v ghc &> /dev/null; then
    echo -e "${RED}❌ GHC is not installed${NC}"
    echo "Please install GHC via ghcup:"
    echo "  curl --proto '=https' --tlsv1.2 -sSf https://get-ghcup.haskell.org | sh"
    echo "  ghcup install ghc 9.6.6"
    echo "  ghcup set ghc 9.6.6"
    exit 1
fi

GHC_VERSION=$(ghc --version | awk '{print $NF}')
echo -e "${GREEN}✓${NC} GHC $GHC_VERSION found"

# Check cabal
if ! command -v cabal &> /dev/null; then
    echo -e "${RED}❌ Cabal is not installed${NC}"
    echo "Please install Cabal via ghcup:"
    echo "  ghcup install cabal"
    exit 1
fi

CABAL_VERSION=$(cabal --version | head -n1 | awk '{print $NF}')
echo -e "${GREEN}✓${NC} Cabal $CABAL_VERSION found"

# Build Haskell rule engine
echo ""
echo "🔧 Building Haskell rule engine..."
cd haskell
cabal update
cabal build exe:kubecheck-rules
HASKELL_BIN=$(cabal list-bin exe:kubecheck-rules)
echo -e "${GREEN}✓${NC} Haskell rule engine built"

# Build Go CLI
echo ""
echo "🔧 Building Go CLI..."
cd ../cmd/kubecheck
go build -o kubecheck
echo -e "${GREEN}✓${NC} Go CLI built"

# Install binaries
echo ""
echo "📦 Installing to /usr/local..."

# Create installation directories
sudo mkdir -p /usr/local/bin
sudo mkdir -p /usr/local/lib/kubecheck

# Install Go binary
echo "  Installing kubecheck CLI..."
sudo cp kubecheck /usr/local/bin/kubecheck
sudo chmod +x /usr/local/bin/kubecheck

# Install Haskell rule engine
echo "  Installing rule engine..."
sudo cp "$HASKELL_BIN" /usr/local/lib/kubecheck/kubecheck-rules
sudo chmod +x /usr/local/lib/kubecheck/kubecheck-rules

echo ""
echo -e "${GREEN}✅ Installation complete!${NC}"
echo ""
echo "Try it out:"
echo "  kubecheck examples/deployment.yaml"
echo ""
echo "To uninstall:"
echo "  ./uninstall.sh"

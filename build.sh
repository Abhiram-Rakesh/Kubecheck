#!/bin/bash
set -e

echo "🔨 Building kubecheck..."

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PROJECT_ROOT=$(pwd)

# Detect OS
if [[ "$OSTYPE" == "linux-gnu"* ]]; then
    OS="linux"
elif [[ "$OSTYPE" == "darwin"* ]]; then
    OS="macos"
else
    echo -e "${RED}❌ Unsupported OS: $OSTYPE${NC}"
    exit 1
fi

echo "📋 Checking prerequisites..."

# Check and offer to install Go
if ! command -v go &> /dev/null; then
    echo -e "${YELLOW}⚠️  Go is not installed${NC}"
    echo ""
    read -p "Install Go 1.21.6 automatically? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "📦 Installing Go..."
        if [ "$OS" = "linux" ]; then
            wget https://go.dev/dl/go1.21.6.linux-amd64.tar.gz -q --show-progress
            sudo rm -rf /usr/local/go
            sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
            rm go1.21.6.linux-amd64.tar.gz
            export PATH=$PATH:/usr/local/go/bin
            if ! grep -q "/usr/local/go/bin" ~/.bashrc; then
                echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
            fi
        elif [ "$OS" = "macos" ]; then
            if ! command -v brew &> /dev/null; then
                echo -e "${RED}❌ Homebrew required. Install from https://brew.sh${NC}"
                exit 1
            fi
            brew install go@1.21
        fi
        echo -e "${GREEN}✓ Go installed${NC}"
    else
        echo -e "${RED}❌ Go is required to build kubecheck${NC}"
        echo "Install manually: https://go.dev/dl/"
        exit 1
    fi
else
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    echo -e "${GREEN}✓${NC} Go $GO_VERSION found"
fi

# Check and offer to install GHC
if ! command -v ghc &> /dev/null; then
    echo -e "${YELLOW}⚠️  GHC is not installed${NC}"
    echo ""
    read -p "Install GHC 9.6.6 via ghcup automatically? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "📦 Installing Haskell toolchain via ghcup..."
        curl --proto '=https' --tlsv1.2 -sSf https://get-ghcup.haskell.org | BOOTSTRAP_HASKELL_NONINTERACTIVE=1 sh
        source "$HOME/.ghcup/env"
        ghcup install ghc 9.6.6
        ghcup install cabal 3.10
        ghcup set ghc 9.6.6
        ghcup set cabal 3.10
        echo -e "${GREEN}✓ Haskell installed${NC}"
    else
        echo -e "${RED}❌ GHC is required to build kubecheck${NC}"
        echo "Install manually: curl --proto '=https' --tlsv1.2 -sSf https://get-ghcup.haskell.org | sh"
        exit 1
    fi
else
    GHC_VERSION=$(ghc --version | awk '{print $NF}')
    echo -e "${GREEN}✓${NC} GHC $GHC_VERSION found"
fi

# Check Cabal
if ! command -v cabal &> /dev/null; then
    echo -e "${RED}❌ Cabal is not installed${NC}"
    echo "Run: ghcup install cabal"
    exit 1
fi

CABAL_VERSION=$(cabal --version | head -n1 | awk '{print $NF}')
echo -e "${GREEN}✓${NC} Cabal $CABAL_VERSION found"

# Check Helm (optional)
if ! command -v helm &> /dev/null; then
    echo -e "${YELLOW}⚠️  Helm is not installed (optional for Helm chart validation)${NC}"
    read -p "Install Helm automatically? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "📦 Installing Helm..."
        if [ "$OS" = "linux" ]; then
            curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
        elif [ "$OS" = "macos" ]; then
            brew install helm
        fi
        echo -e "${GREEN}✓ Helm installed${NC}"
    else
        echo -e "${BLUE}ℹ️  Skipping Helm (you can install later for Helm chart validation)${NC}"
    fi
fi

# Build Haskell rule engine
echo ""
echo "🔧 Building Haskell rule engine..."
cd "$PROJECT_ROOT/haskell"
cabal update
cabal build exe:kubecheck-rules
HASKELL_BIN=$(cabal list-bin exe:kubecheck-rules)
echo -e "${GREEN}✓${NC} Haskell rule engine built"

# Build Go CLI
echo ""
echo "🔧 Building Go CLI..."
cd "$PROJECT_ROOT/cmd/kubecheck"
go mod tidy
go build -o kubecheck
echo -e "${GREEN}✓${NC} Go CLI built"

# Install binaries
echo ""
echo "📦 Installing to /usr/local..."

sudo mkdir -p /usr/local/bin
sudo mkdir -p /usr/local/lib/kubecheck

echo "  Installing kubecheck CLI..."
sudo cp kubecheck /usr/local/bin/kubecheck
sudo chmod +x /usr/local/bin/kubecheck

echo "  Installing rule engine..."
sudo cp "$HASKELL_BIN" /usr/local/lib/kubecheck/kubecheck-rules
sudo chmod +x /usr/local/lib/kubecheck/kubecheck-rules

echo ""
echo -e "${GREEN}✅ Installation complete!${NC}"
echo ""
echo "Try it out:"
echo "  kubecheck examples/deployment.yaml"
echo ""
echo "To uninstall kubecheck:"
echo "  ./uninstall.sh"

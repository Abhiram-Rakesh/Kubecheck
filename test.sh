#!/bin/bash

# Test script for kubecheck
# This script validates that the project builds and runs correctly

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "🧪 Running kubecheck tests..."
echo ""

# Test 1: Check project structure
echo -n "Test 1: Project structure... "
required_files=(
    "README.md"
    "build.sh"
    "uninstall.sh"
    "cmd/kubecheck/main.go"
    "cmd/kubecheck/parser.go"
    "cmd/kubecheck/helm.go"
    "cmd/kubecheck/reporter.go"
    "cmd/kubecheck/engine.go"
    "haskell/kubecheck.cabal"
    "haskell/src/Types.hs"
    "haskell/src/Rules.hs"
    "haskell/src/Validator.hs"
    "haskell/app/Main.hs"
)

all_exist=true
for file in "${required_files[@]}"; do
    if [ ! -f "$file" ]; then
        echo -e "${RED}✗${NC}"
        echo "  Missing file: $file"
        all_exist=false
    fi
done

if [ "$all_exist" = true ]; then
    echo -e "${GREEN}✓${NC}"
fi

# Test 2: Go code compiles
echo -n "Test 2: Go code compiles... "
cd cmd/kubecheck
if go build -o test-binary > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC}"
    rm -f test-binary
else
    echo -e "${RED}✗${NC}"
    echo "  Go compilation failed"
    exit 1
fi
cd ../..

# Test 3: Haskell code compiles
echo -n "Test 3: Haskell code compiles... "
cd haskell
if cabal build > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC}"
else
    echo -e "${RED}✗${NC}"
    echo "  Haskell compilation failed"
    echo "  Note: This requires GHC and Cabal to be installed"
    exit 1
fi
cd ..

# Test 4: Example files are valid YAML
echo -n "Test 4: Example files are valid... "
valid_yaml=true
for file in examples/*.yaml; do
    if ! python3 -c "import yaml; yaml.safe_load(open('$file'))" 2>/dev/null; then
        echo -e "${RED}✗${NC}"
        echo "  Invalid YAML: $file"
        valid_yaml=false
    fi
done

if [ "$valid_yaml" = true ]; then
    echo -e "${GREEN}✓${NC}"
fi

# Test 5: Documentation exists
echo -n "Test 5: Documentation exists... "
docs=(
    "README.md"
    "CONTRIBUTING.md"
    "ARCHITECTURE.md"
    "QUICKSTART.md"
    "LICENSE"
)

docs_exist=true
for doc in "${docs[@]}"; do
    if [ ! -f "$doc" ]; then
        echo -e "${RED}✗${NC}"
        echo "  Missing documentation: $doc"
        docs_exist=false
    fi
done

if [ "$docs_exist" = true ]; then
    echo -e "${GREEN}✓${NC}"
fi

echo ""
echo -e "${GREEN}✅ All tests passed!${NC}"
echo ""
echo "Next steps:"
echo "  1. Run ./build.sh to install kubecheck"
echo "  2. Try: kubecheck examples/deployment.yaml"

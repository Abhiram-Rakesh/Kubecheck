# Contributing to kubecheck

Thank you for your interest in contributing to kubecheck! This guide will help you get started.

## Architecture Overview

kubecheck uses a hybrid architecture:

```
┌──────────────────────────────────────────┐
│  Go CLI (cmd/kubecheck)                  │
│  - Argument parsing                      │
│  - File discovery (YAML, directories)    │
│  - Helm template rendering               │
│  - JSON serialization                    │
│  - Report formatting                     │
└────────────┬─────────────────────────────┘
             │ JSON over stdin/stdout
             ▼
┌──────────────────────────────────────────┐
│  Haskell Rule Engine (haskell/)          │
│  - Type-safe resource parsing            │
│  - Pure functional validation            │
│  - Declarative rule definitions          │
│  - Violation generation                  │
└──────────────────────────────────────────┘
```

## Development Setup

### Prerequisites

- Go ≥ 1.21
- GHC 9.6.x (via ghcup)
- Cabal ≥ 3.0
- Helm (optional, for Helm chart testing)

### Installing Dependencies

```bash
# Install ghcup (Haskell toolchain manager)
curl --proto '=https' --tlsv1.2 -sSf https://get-ghcup.haskell.org | sh

# Install GHC and Cabal
ghcup install ghc 9.6.6
ghcup install cabal
ghcup set ghc 9.6.6

# Verify installations
go version
ghc --version
cabal --version
```

### Local Development Build

```bash
# Build Go CLI
cd cmd/kubecheck
go build -o kubecheck
cd ../..

# Build Haskell rule engine
cd haskell
cabal update
cabal build
cd ..

# Test locally (without installing)
./cmd/kubecheck/kubecheck examples/deployment.yaml
```

## Adding New Rules

New validation rules should be added to the Haskell rule engine for type safety and functional correctness.

### Step 1: Define the Rule

Edit `haskell/src/Rules.hs`:

```haskell
-- | Rule: Check for host network usage
-- Severity: ERROR
-- Rationale: Host network bypasses network policies
checkHostNetwork :: Container -> [Violation]
checkHostNetwork container =
    case containerHostNetwork container of
        Just True -> [Violation SeverityError msg "no-host-network"]
        _         -> []
  where
    msg = "Container '" <> containerName container
          <> "' uses host network (hostNetwork: true)"
```

### Step 2: Apply the Rule

Edit `haskell/src/Validator.hs`:

```haskell
validateContainer :: Container -> [Violation]
validateContainer container =
    concat [ checkNoLatestImage container
           , checkRequireResources container
           , checkNoRootContainers container
           , checkHostNetwork container  -- Add your new rule
           ]
```

### Step 3: Update Types (if needed)

If your rule requires new fields, update `haskell/src/Types.hs`:

```haskell
data Container = Container
    { containerName           :: Text
    , containerImage          :: Text
    , containerSecurityContext :: Maybe SecurityContext
    , containerResources      :: Maybe Resources
    , containerHostNetwork    :: Maybe Bool  -- Add new field
    } deriving (Eq, Show, Generic)
```

### Step 4: Test Your Rule

Create a test case in `examples/`:

```yaml
# examples/host-network.yaml
apiVersion: v1
kind: Pod
metadata:
  name: test-host-network
spec:
  hostNetwork: true
  containers:
    - name: app
      image: nginx:1.21
```

Test it:

```bash
./build.sh
kubecheck examples/host-network.yaml
```

## Testing

### Go Tests

```bash
cd cmd/kubecheck
go test -v ./...
```

### Haskell Tests

```bash
cd haskell
cabal test
```

### Integration Tests

```bash
# Test single file
kubecheck examples/deployment.yaml

# Test directory
kubecheck examples/

# Test Helm chart
kubecheck examples/helm-chart/

# Test stdin
cat examples/pod.yaml | kubecheck -
```

## Code Style

### Go Code Style

- Follow standard Go conventions
- Use `gofmt` for formatting
- Run `go vet` before committing
- Keep functions focused and testable

### Haskell Code Style

- Use Haskell2010 standard
- Enable `-Wall` and treat warnings as errors
- Use type signatures for all top-level functions
- Prefer pure functions over IO
- Use meaningful names (no single-letter variables except in small scopes)

## Pull Request Process

1. **Fork the repository**
2. **Create a feature branch**
   ```bash
   git checkout -b feature/my-new-rule
   ```
3. **Make your changes**
   - Add your rule to `haskell/src/Rules.hs`
   - Update `haskell/src/Validator.hs`
   - Add test cases to `examples/`
   - Update documentation if needed
4. **Test thoroughly**
   ```bash
   ./build.sh
   kubecheck examples/
   ```
5. **Commit with clear messages**
   ```bash
   git commit -m "feat: add host network validation rule"
   ```
6. **Push and create PR**
   ```bash
   git push origin feature/my-new-rule
   ```

## Rule Design Guidelines

### Severity Levels

- **ERROR**: Production-critical violations
  - Security issues (running as root, privileged containers)
  - Non-deterministic configurations (`:latest` tags)
  - Critical misconfigurations

- **WARN**: Best practice violations
  - Missing resource limits
  - Missing probes
  - Suboptimal configurations

- **OK**: All checks passed

### Rule Characteristics

Good rules should be:

1. **Deterministic** - Same input always produces same output
2. **Actionable** - Clear guidance on how to fix
3. **Documented** - Include rationale in comments
4. **Testable** - Easy to create test cases
5. **Focused** - One rule checks one thing

### Rule Template

```haskell
-- | Rule: <Brief description>
-- Severity: <ERROR|WARN>
-- Rationale: <Why this rule exists>
checkRuleName :: Container -> [Violation]
checkRuleName container =
    if condition
        then [Violation severity message "rule-name"]
        else []
  where
    message = "Clear, actionable error message"
    condition = -- Your validation logic
```

## Debugging

### Debug Go CLI

```bash
cd cmd/kubecheck
go run . -v examples/deployment.yaml
```

### Debug Haskell Rule Engine

```bash
cd haskell
echo '{"apiVersion":"v1","kind":"Pod",...}' | cabal run kubecheck-rules
```

### Common Issues

1. **Rule engine not found**
   - Ensure `/usr/local/lib/kubecheck/kubecheck-rules` exists
   - Re-run `./build.sh`

2. **JSON parsing errors**
   - Verify JSON structure matches `Types.hs` definitions
   - Check that all required fields are present

3. **Haskell compilation errors**
   - Run `cabal clean && cabal build`
   - Ensure GHC version is 9.6.x

## Resources

- [Kubernetes Best Practices](https://kubernetes.io/docs/concepts/configuration/overview/)
- [NSA Kubernetes Hardening Guide](https://www.nsa.gov/Press-Room/News-Highlights/Article/Article/2716980/)
- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)
- [Haskell Style Guide](https://kowainik.github.io/posts/2019-02-06-style-guide)

## Community

- GitHub Issues: Bug reports and feature requests
- GitHub Discussions: Questions and ideas

---

**Thank you for contributing to kubecheck!**

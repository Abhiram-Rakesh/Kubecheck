# kubecheck

**Production-grade Kubernetes static analysis CLI tool**

A local CLI tool that validates Kubernetes YAML files against production best practices. Designed for CI/CD pipelines, pre-commit hooks, and local developer validation.

## Overview

`kubecheck` is a static analysis tool that validates Kubernetes manifests without connecting to a cluster. It combines:

- **Go** - CLI orchestration, filesystem traversal, Helm integration
- **Haskell** - Pure functional rule engine for declarative policy validation

## Architecture

```
┌─────────────────────────────────────────┐
│          kubecheck CLI (Go)             │
│  - Argument parsing                     │
│  - File discovery & YAML parsing        │
│  - Helm template rendering              │
│  - Orchestration & reporting            │
└──────────────┬──────────────────────────┘
               │ JSON
               ▼
┌─────────────────────────────────────────┐
│     Rule Engine (Haskell)               │
│  - Pure functional validation           │
│  - Declarative rule definitions         │
│  - Severity classification              │
└─────────────────────────────────────────┘
```

## Features

### Input Support

- Single Kubernetes YAML files
- Directories (recursive scanning)
- Multi-document YAML files (`---` separated)
- Helm charts (via `helm template`)
- Stdin piping

### Validation Rules

Current production best practices:

| Rule                 | Severity | Description                            |
| -------------------- | -------- | -------------------------------------- |
| `no-latest-image`    | ERROR    | Disallow `image: latest` tags          |
| `require-resources`  | WARN     | Require CPU/memory requests and limits |
| `no-root-containers` | ERROR    | Detect containers running as root      |

Rules are extensible via the Haskell rule engine.

### Exit Codes

```bash
0 - OK    (all checks passed)
1 - WARN  (warnings found)
2 - ERROR (errors found)
```

The CLI exits with the highest severity found, making it CI-friendly.

## Installation

### Prerequisites

- Go ≥ 1.21
- GHC 9.6.x (installed via ghcup, no system GHC required)
- Helm (optional, for Helm chart validation)

### Install

```bash
git clone <repository-url>
cd kubecheck
chmod +x *.sh
./build.sh
```

This installs:

- `kubecheck` binary to `/usr/local/bin`
- Haskell rule engine to `/usr/local/lib/kubecheck`

### Uninstall

```bash
./uninstall.sh
```

## Usage

### Basic Usage

```bash
# Validate a single file
kubecheck deployment.yaml

# Validate a directory (recursive)
kubecheck k8s/

# Validate a Helm chart
kubecheck ./my-chart/

# Pipe from stdin
helm template ./my-chart | kubecheck -

# Verbose output
kubecheck -v deployment.yaml
```

### Examples

**Single file validation:**

```bash
$ kubecheck examples/deployment.yaml
ERROR: examples/deployment.yaml
  Container 'nginx' uses 'latest' image tag
  Container 'nginx' missing resource requests/limits
  Container 'nginx' running as root (securityContext.runAsNonRoot not set)
Exit code: 2
```

**Directory validation:**

```bash
$ kubecheck k8s/
OK: k8s/service.yaml
WARN: k8s/deployment.yaml
  Container 'app' missing resource limits
ERROR: k8s/pod.yaml
  Container 'debug' uses 'latest' image tag
Exit code: 2
```

**Helm chart validation:**

```bash
$ kubecheck ./charts/myapp/
Rendering Helm chart...
ERROR: charts/myapp/templates/deployment.yaml
  Container 'myapp' uses 'latest' image tag
Exit code: 2
```

**CI/CD Integration:**

```yaml
# .github/workflows/validate.yml
- name: Validate Kubernetes manifests
  run: |
    ./kubecheck k8s/
  # Fails the build on ERROR (exit code 2)
```

## Testing

```bash
# Run Go tests
cd cmd/kubecheck
go test -v

# Run Haskell tests
cd haskell
cabal test
```

## Contributing

### Adding New Rules

1. Define the rule in `haskell/src/Rules.hs`
2. Add validation logic in `haskell/src/Validator.hs`
3. Rebuild and test

Example:

```haskell
-- In Rules.hs
checkHostNetwork :: Container -> [Violation]
checkHostNetwork container =
  if containerHostNetwork container
    then [Violation Error "Container uses host network"]
    else []
```

## Acknowledgments

Built with best practices from:

- [Kubernetes Production Best Practices](https://learnk8s.io/production-best-practices)
- [NSA Kubernetes Hardening Guide](https://www.nsa.gov/Press-Room/News-Highlights/Article/Article/2716980/)
- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)

---

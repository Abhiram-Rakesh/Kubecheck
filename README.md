# kubecheck

**Production-grade Kubernetes static analysis CLI tool**

A local CLI tool that validates Kubernetes YAML files against production best practices. Designed for CI/CD pipelines, pre-commit hooks, and local developer validation.

## Overview

`kubecheck` is a static analysis tool that validates Kubernetes manifests without connecting to a cluster. Built with Go and featuring a YAML-configurable rule system for maximum flexibility.

## Architecture

```
┌─────────────────────────────────────────┐
│          kubecheck CLI (Go)             │
│  - Argument parsing                     │
│  - File discovery & YAML parsing        │
│  - Helm template rendering              │
│  - Orchestration & reporting            │
│  - YAML-configurable rule engine        │
└─────────────────────────────────────────┘
```

## Features

### Input Support

- ✅ Single Kubernetes YAML files
- ✅ Directories (recursive scanning)
- ✅ Multi-document YAML files (`---` separated)
- ✅ Helm charts (via `helm template`)
- ✅ Stdin piping

### YAML-Configurable Rules

Organizations can define custom validation rules via YAML configuration:

```yaml
rules:
  - name: no-latest-image
    severity: ERROR
    conditions:
      - image_tag_equals:latest
    message: "Container '{container}' uses 'latest' image tag"
```

See [CONFIG.md](CONFIG.md) for complete documentation.

### Default Validation Rules

| Rule                        | Severity | Description                       |
| --------------------------- | -------- | --------------------------------- |
| `no-latest-image`           | ERROR    | Disallow `image: latest` tags     |
| `require-resource-requests` | WARN     | Require CPU/memory requests       |
| `require-resource-limits`   | WARN     | Require CPU/memory limits         |
| `no-root-containers`        | ERROR    | Detect containers running as root |

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
- Helm (optional, for Helm chart validation)

### Install

```bash
git clone https://github.com/Abhiram-Rakesh/kubecheck.git
cd kubecheck
chmod +x *.sh
./build.sh
```

This installs:

- `kubecheck` binary to `/usr/local/bin`

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

# Use custom config
kubecheck --config my-rules.yaml deployment.yaml
```

### Configuration

kubecheck looks for configuration files in:

1. `./kubecheck.yaml` (current directory)
2. `./kubecheck.yml` (current directory)
3. `~/.kubecheck/config.yaml` (home directory)
4. Built-in defaults (if no config found)

Create a custom config:

```yaml
# kubecheck.yaml
rules:
  - name: require-company-registry
    description: All images must use company registry
    severity: ERROR
    type: image
    conditions:
      - image_not_from_registry:registry.company.com
    message: "Container '{container}' uses external registry"
    help: "use images from registry.company.com"
```

See [CONFIG.md](CONFIG.md) for complete configuration guide.

### Examples

**Single file validation:**

```bash
$ kubecheck examples/deployment.yaml

  ● File: examples/deployment.yaml
  ┌─ Deployment: nginx-deployment ──────────────────────────┐
  │  ✖  Security Violation
  │     Container 'nginx' uses 'latest' image tag
  │     ▲─── use a specific version or digest
  │
  │  ⚠  Resource Hygiene
  │     Container 'nginx' missing resource requests/limits
  └────────────────────────────────────── [ 1 errors | 1 warns ]

  Summary ➔ 1 file checked. 2 violations found.
```

**Directory validation:**

```bash
$ kubecheck k8s/

  🔍 Scanning directory: ./k8s/
  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  ✔  k8s/service.yaml ........................... PASSED
  ⚠  k8s/deployment.yaml ........................ 1 WARN
     └─ [api-server] Container 'api' missing resource limits
  ✖  k8s/cronjob.yaml ........................... 1 ERR
     └─ [backup] Container 'backup' uses 'latest' image tag

  ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  Summary ➔ 3 files checked
  Result  ➔ 1 OK  |  1 Warning  |  1 Error
  Status  ➔ FAILED Exit code: 2
```

## Contributing

### Adding New Rules

Rules are defined in YAML configuration files. To add new condition types, edit `cmd/kubecheck/rule-engine.go`:

```go
// Add new condition to checkCondition switch
case "my_new_condition":
    return checkMyCondition(container)
```

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guide.

## Documentation

- [CONFIG.md](CONFIG.md) - Configuration guide
- [QUICKSTART.md](QUICKSTART.md) - Get started in 5 minutes
- [ARCHITECTURE.md](ARCHITECTURE.md) - System design
- [CONTRIBUTING.md](CONTRIBUTING.md) - How to contribute
- [EXAMPLES.md](EXAMPLES.md) - Real-world usage examples

## Acknowledgments

Built with best practices from:

- [Kubernetes Production Best Practices](https://learnk8s.io/production-best-practices)
- [NSA Kubernetes Hardening Guide](https://www.nsa.gov/Press-Room/News-Highlights/Article/Article/2716980/)
- [CIS Kubernetes Benchmark](https://www.cisecurity.org/benchmark/kubernetes)

---

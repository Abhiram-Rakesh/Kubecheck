# Quickstart Guide

Get up and running with `kubecheck` in 5 minutes.

## ⚡ Quick Install

```bash
# Clone the repository
git clone https://github.com/Abhiram-Rakesh/kubecheck.git
cd kubecheck

# Make scripts executable
chmod +x *.sh

# Build and install
./build.sh
```

**Prerequisites:** Go ≥ 1.21, Helm (optional)

## Basic Usage

### Validate a Single File

```bash
kubecheck examples/deployment.yaml
```

**Output:**

```
  ● File: examples/deployment.yaml
  ┌─ Deployment: nginx-deployment ──────────────────────┐
  │  ✖  Security Violation
  │     Container 'nginx' uses 'latest' image tag
  │     ▲─── use a specific version or digest
  │
  │  ⚠  Resource Hygiene
  │     Container 'nginx' missing resource requests
  └──────────────────────────────────── [ 1 errors | 1 warns ]
```

**Exit code:** 2 (ERROR level found)

### Validate a Directory

```bash
kubecheck k8s/
```

Recursively scans all `.yaml` and `.yml` files.

### Validate a Helm Chart

```bash
kubecheck ./my-chart/
```

Automatically detects Helm charts (looks for `Chart.yaml`) and renders templates.

### Pipe from stdin

```bash
helm template ./my-chart | kubecheck -
```

Or:

```bash
kubectl get deployment nginx -o yaml | kubecheck -
```

## Understanding Exit Codes

| Code | Meaning | Description                    |
| ---- | ------- | ------------------------------ |
| 0    | OK      | All checks passed              |
| 1    | WARN    | Warnings found, but no errors  |
| 2    | ERROR   | Errors found (fails CI builds) |

The CLI exits with the **highest severity** found across all files.

## Verbose Mode

Get detailed output for all files, including those that pass:

```bash
kubecheck -v k8s/
```

## What kubecheck Checks

### Default Rules

If no config file is found, kubecheck uses these built-in rules:

**❌ Errors (Exit Code 2)**

1. **No `:latest` image tags**
2. **Containers must not run as root**

**⚠️ Warnings (Exit Code 1)**

1. **Resource requests should be set**
2. **Resource limits should be set**

## Custom Configuration

Create a `kubecheck.yaml` file in your project:

```yaml
rules:
  - name: no-latest-image
    severity: ERROR
    type: image
    conditions:
      - image_tag_equals:latest
    message: "Container '{container}' uses latest tag"
```

kubecheck will automatically find and use it. See [CONFIG.md](CONFIG.md) for details.

## Common Workflows

### Pre-Commit Hook

`.git/hooks/pre-commit`:

```bash
#!/bin/bash
kubecheck k8s/
```

### CI/CD Pipeline

**GitHub Actions:**

```yaml
- name: Validate Kubernetes Manifests
  run: kubecheck k8s/
```

**GitLab CI:**

```yaml
validate:
  script:
    - kubecheck k8s/
```

### Local Development

```bash
# Check before committing
kubecheck k8s/

# Verbose output for debugging
kubecheck -v deployment.yaml

# Check Helm chart
kubecheck ./charts/myapp/

# Check specific rendered values
helm template ./charts/myapp -f prod-values.yaml | kubecheck -
```

## Example Manifests

The `examples/` directory contains sample manifests:

```bash
# Bad practices (will fail)
kubecheck examples/deployment.yaml
kubecheck examples/pod.yaml

# Good practices (will pass)
kubecheck examples/deployment-good.yaml

# Multi-document YAML
kubecheck examples/multi-doc.yaml

# Helm chart
kubecheck examples/helm-chart/
```

## Troubleshooting

### "kubecheck: command not found"

Make sure `/usr/local/bin` is in your `PATH`:

```bash
export PATH="/usr/local/bin:$PATH"
```

Or add to `~/.bashrc` or `~/.zshrc`.

### "helm template failed"

Ensure Helm is installed:

```bash
brew install helm  # macOS
# or
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

## Uninstall

```bash
./uninstall.sh
```

## Next Steps

1. **Read [CONFIG.md](CONFIG.md)** to customize rules for your organization
2. **Explore [ARCHITECTURE.md](ARCHITECTURE.md)** to understand how it works
3. **Check [EXAMPLES.md](EXAMPLES.md)** for CI/CD integration patterns

---

**Happy validating!**

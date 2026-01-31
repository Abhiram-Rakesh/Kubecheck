# Quickstart Guide

Get up and running with `kubecheck` in 5 minutes.

## Quick Install

```bash
# Clone the repository
git clone https://github.com/Abhiram-Rakesh/Kubecheck.git
cd kubecheck

# Make scripts executable
chmod +x *.sh

# Build and install
./build.sh
```

**Prerequisites:** Go ≥ 1.21, GHC 9.6.x (installed via ghcup)

## Basic Usage

### Validate a Single File

```bash
kubecheck examples/deployment.yaml
```

**Output:**

```
ERROR: examples/deployment.yaml
  Resource: Deployment/nginx-deployment
  [ERROR] Container 'nginx' uses 'latest' image tag
  [WARN] Container 'nginx' missing resource requests/limits
  [ERROR] Container 'nginx' missing securityContext (should set runAsNonRoot: true)

─────────────────────────────────────
Summary: 1 files checked
  ✗ 1 with errors
Total violations: 3
─────────────────────────────────────
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

### Errors (Exit Code 2)

1. **No `:latest` image tags**

   ```yaml
   # ❌ Bad
   image: nginx:latest
   image: nginx

   # ✅ Good
   image: nginx:1.21
   image: nginx:1.21.0-alpine
   ```

2. **Containers must not run as root**

   ```yaml
   # ❌ Bad
   spec:
     containers:
     - name: app
       image: nginx:1.21
       # Missing securityContext

   # ✅ Good
   spec:
     containers:
     - name: app
       image: nginx:1.21
       securityContext:
         runAsNonRoot: true
         runAsUser: 1000
   ```

### ⚠️ Warnings (Exit Code 1)

1. **Resource requests and limits should be set**

   ```yaml
   # ⚠️ Warning
   spec:
     containers:
     - name: app
       image: nginx:1.21
       # Missing resources

   # ✅ Good
   spec:
     containers:
     - name: app
       image: nginx:1.21
       resources:
         requests:
           cpu: "100m"
           memory: "128Mi"
         limits:
           cpu: "500m"
           memory: "512Mi"
   ```

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
name: Validate Kubernetes Manifests

on: [push, pull_request]

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup kubecheck
        run: |
          git clone https://github.com/your-org/kubecheck
          cd kubecheck
          ./build.sh

      - name: Validate manifests
        run: kubecheck k8s/
```

**GitLab CI:**

```yaml
validate:
  stage: test
  script:
    - ./kubecheck k8s/
  artifacts:
    when: always
    reports:
      junit: report.xml
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

### "Failed to parse JSON input"

The Haskell rule engine expects valid Kubernetes YAML. Check that your manifests have:

- `apiVersion`
- `kind`
- `metadata`

## Next Steps

1. **Read the [Contributing Guide](CONTRIBUTING.md)** to add your own rules
2. **Explore the [Architecture](ARCHITECTURE.md)** to understand how it works
3. **Check the [README](README.md)** for comprehensive documentation

## Uninstall

```bash
./uninstall.sh
```

Removes all installed binaries and libraries.

---
